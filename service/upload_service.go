package service

import (
	"DiTing-Go/dal/model"
	"DiTing-Go/domain/dto"
	"DiTing-Go/domain/enum"
	voResp "DiTing-Go/domain/vo/resp"
	"DiTing-Go/global"
	"DiTing-Go/pkg/domain/vo/resp"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/minio/minio-go/v7"
	"strconv"
	"time"
)

// GetPreSigned 签发url
//
//	@Summary	签发url
//	@Produce	json
//	@Param		roomId	query		int64				true	"房间ID"
//	@Param		fileName	query		string				true	"文件名"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/file/getPreSigned [get]
func GetPreSigned(c *gin.Context) {
	// 获取上下文中的用户 ID
	uid := c.GetInt64("uid")
	// 创建一个新的上下文
	ctx := context.Background()

	// 获取请求中的 roomId 参数
	roomIdStr, found := c.GetQuery("roomId")
	if !found {
		// 记录错误并返回错误响应
		global.Logger.Errorf("参数错误 %s", roomIdStr)
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	// 将 roomId 从字符串转换为 int64
	roomId, err := strconv.ParseInt(roomIdStr, 10, 64)
	if err != nil {
		// 记录错误并返回错误响应
		global.Logger.Errorf("参数错误 %s", roomId)
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}

	// 获取请求中的 fileName 参数
	fileName, found := c.GetQuery("fileName")
	if !found {
		// 记录错误并返回错误响应
		global.Logger.Errorf("参数错误 %s", roomIdStr)
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	// 构造文件名：时间戳 + 用户ID + 文件名
	// 按天创建桶
	timeStr := time.Now().Format("2006-01-02")
	fileName = fmt.Sprintf("%s/%d/%s", timeStr, uid, fileName)

	// 创建一个新的 POST 签名策略对象
	policy := minio.NewPostPolicy()
	// 设置目标存储桶的名称
	if err := policy.SetBucket("diting"); err != nil {
		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}
	// 设置上传的目标文件路径
	if err := policy.SetKey(fileName); err != nil {
		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}
	// 设置策略的过期时间为1天
	if err := policy.SetExpires(time.Now().UTC().AddDate(0, 0, 1)); err != nil {
		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}
	// 生成预签名的 URL 和表单数据
	url, formData, err := global.MinioClient.PresignedPostPolicy(ctx, policy)
	if err != nil {
		global.Logger.Errorf("创建policy失败 %s", roomIdStr)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}
	// 构建响应数据
	preSignedResp := voResp.PreSignedResp{
		Url:    url.String(),
		Policy: formData,
	}

	// 开始数据库事务
	tx := global.Query.Begin()
	// 在事务上下文中操作消息表
	messageTx := tx.Message.WithContext(ctx)
	base := dto.MessageBaseDto{
		Url:  url.String(),
		Size: -1, // 文件大小未知，初始化为 -1
		Name: fileName,
	}
	// 构造图片消息的额外信息
	extra := dto.ImgMessageDto{
		MessageBaseDto: base,
		Width:          -1, // TODO: 前端传入宽度
		Height:         -1, // TODO: 前端传入高度
	}
	// 序列化消息的额外信息
	jsonStr, err := json.Marshal(extra)
	if err != nil {
		// 序列化失败时回滚事务
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err)
		}
		global.Logger.Errorf("json序列化失败 %s", err)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}
	// 创建新的消息记录
	newMsg := model.Message{
		FromUID:      uid,
		RoomID:       roomId,
		Content:      "[图片]",
		DeleteStatus: 0,
		Type:         3,
		Extra:        string(jsonStr),
	}
	// 插入消息到数据库
	if err := messageTx.Create(&newMsg); err != nil {
		// 数据库插入失败时回滚事务
		if err := tx.Rollback(); err != nil {
			global.Logger.Errorf("事务回滚失败 %s", err)
		}
		global.Logger.Errorf("数据库插入失败 %s", err)
		resp.ErrorResponse(c, "获取签名失败，请稍后再试")
		c.Abort()
		return
	}

	// 发布新消息事件
	global.Bus.Publish(enum.NewMessageEvent, newMsg)

	// 返回成功响应
	resp.SuccessResponse(c, preSignedResp)
}
