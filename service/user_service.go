package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/dal/query"
	domainEnum "DiTing-Go/domain/enum"
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	pkgResp "DiTing-Go/pkg/domain/vo/resp"
	_ "DiTing-Go/pkg/setting"
	"DiTing-Go/pkg/utils"
	"DiTing-Go/service/adapter"
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var q *query.Query = global.Query

// RegisterService 用户注册
// RegisterService 是用户注册服务的实现
func RegisterService(userReq req.UserRegisterReq) (pkgResp.ResponseData, error) {
	// 创建一个新的上下文
	ctx := context.Background()
	// 获取全局的用户查询对象
	user := global.Query.User
	// 将上下文添加到用户查询对象中
	userQ := user.WithContext(ctx)
	// 定义一个函数，该函数将在数据库中查找与请求中的用户名匹配的用户
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(userReq.Username)).First()
	}
	// 创建一个用户模型对象
	userR := model.User{}
	// 生成一个缓存键，该键基于请求中的用户名
	key := fmt.Sprintf(domainEnum.UserCacheByName, userReq.Username)
	// 尝试从缓存或数据库中获取用户数据
	err := utils.GetData(key, &userR, fun)
	// 如果没有错误，说明找到了匹配的用户，因此返回一个错误响应
	if err == nil {
		return pkgResp.ErrorResponseData("用户名已存在"), errors.New("Business Error")
	}
	// 如果错误不是记录未找到的错误，说明查询过程中出现了问题，因此返回一个错误响应
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		global.Logger.Errorf("查询数据失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 创建一个新的用户对象，该对象的属性基于请求中的数据
	newUser := model.User{
		Name:     userReq.Username,
		Password: userReq.Password,
		IPInfo:   "{}",
	}
	// 尝试在数据库中创建新的用户对象，如果出现错误，返回一个错误响应
	if err := userQ.Omit(user.OpenID).Create(&newUser); err != nil {
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 如果一切顺利，返回一个成功的响应
	return pkgResp.SuccessResponseDataWithMsg("注册成功"), nil
}

// LoginService 用户登录
func LoginService(loginReq req.UserLoginReq) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := query.User
	userQ := user.WithContext(ctx)
	// 查数据库
	// 检查密码是否正确
	fun := func() (interface{}, error) {
		return userQ.Where(user.Name.Eq(loginReq.UserName), user.Password.Eq(loginReq.Password)).First()
	}
	userR := model.User{}
	key := fmt.Sprintf(domainEnum.UserCacheByName, loginReq.UserName)
	err := utils.GetData(key, &userR, fun)
	if err != nil {
		global.Logger.Errorf("查询数据失败: %v", err)
		return pkgResp.ErrorResponseData("用户名或密码错误"), errors.New("Business Error")
	}
	//生成jwt
	token, err := utils.GenerateToken(userR.ID)
	if err != nil {
		global.Logger.Errorf("生成jwt失败 %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	// 发送用户登录事件
	userByte, err := json.Marshal(userR)
	if err != nil {
		global.Logger.Errorf("json序列化失败 %v", err)
	}
	msg := &primitive.Message{
		Topic: domainEnum.UserLoginTopic,
		Body:  userByte,
	}
	_, _ = global.RocketProducer.SendSync(ctx, msg)
	userResp := resp.UserLoginResp{
		Token:  token,
		Uid:    userR.ID,
		Name:   userR.Name,
		Avatar: userR.Avatar,
	}
	return pkgResp.SuccessResponseData(userResp), nil
}

func GetUserInfoByNameService(uid int64, name string) (pkgResp.ResponseData, error) {
	ctx := context.Background()
	user := global.Query.User
	userQ := user.WithContext(ctx)
	userRList, err := userQ.Where(user.Name.Like(name + "%")).Limit(5).Find()
	if err != nil {
		global.Logger.Errorf("查询用户数据失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	uidList := make([]int64, 0)
	for _, userR := range userRList {
		uidList = append(uidList, userR.ID)
	}
	//	搜索好友关系
	userApply := global.Query.UserApply
	userApplyQ := userApply.WithContext(ctx)
	applyList, err := userApplyQ.Where(userApply.UID.Eq(uid), userApply.TargetID.In(uidList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询好友关系失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	//	查询好友关系
	userFriend := global.Query.UserFriend
	userFriendQ := userFriend.WithContext(ctx)
	friendList, err := userFriendQ.Where(userFriend.UID.Eq(uid), userFriend.FriendUID.In(uidList...)).Find()
	if err != nil {
		global.Logger.Errorf("查询好友关系失败: %v", err)
		return pkgResp.ErrorResponseData("系统繁忙，请稍后再试~"), errors.New("Business Error")
	}
	userRespList := adapter.BuildUserInfoByNameResp(userRList, applyList, friendList)
	return pkgResp.SuccessResponseData(userRespList), nil
}
