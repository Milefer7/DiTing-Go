package routes

import (
	"DiTing-Go/controller"
	_ "DiTing-Go/docs"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/pkg/middleware"
	"DiTing-Go/service"
	"DiTing-Go/websocket/global"
	websocketService "DiTing-Go/websocket/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"net/http"
)

// InitRouter 初始化路由
func InitRouter() {
	go initWebSocket()
	initGin()
}

// 初始化websocket
func initWebSocket() {
	// 使用 http.HandleFunc 注册一个 HTTP 处理程序，用于处理 "/websocket" 路径上的请求
	// websocketService.Connect 是一个处理函数，用于处理来自客户端的 WebSocket 连接请求
	http.HandleFunc("/websocket", websocketService.Connect)

	// 启动 HTTP 服务器，监听 localhost:5001 地址
	// 任何访问该地址的 HTTP 请求都会被 http.DefaultServeMux 处理
	log.Fatal(http.ListenAndServe("localhost:5001", nil))
}

// 初始化gin
func initGin() {
	router := gin.Default()
	router.Use(middleware.LoggerToFile())
	router.Use(middleware.Cors())
	// 添加swagger访问路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 不需要身份验证的路由
	apiPublic := router.Group("/api/public")
	{
		//注册
		apiPublic.POST("/register", controller.RegisterController)
		//登录
		apiPublic.POST("/login", controller.LoginController)
	}

	apiUser := router.Group("/api/user")
	apiUser.Use(middleware.JWT())
	{
		//添加好友
		apiUser.POST("/add", controller.ApplyFriendController)
		//删除好友
		apiUser.DELETE("/delete/", controller.DeleteFriendController)
		//同意好友申请
		apiUser.PUT("/agree", controller.AgreeFriendController)
		//获取好友申请列表
		apiUser.GET("/getApplyList", controller.GetUserApplyController)
		//获取好友列表
		apiUser.GET("/getFriendList", controller.GetFriendListController)
		// 判断是否是好友
		apiUser.GET("/isFriend/:friendUid", controller.IsFriendController)
		//好友申请未读数量
		apiUser.GET("/unreadApplyNum", controller.UnreadApplyNumController)
		//根据好友昵称搜索好友
		apiUser.GET("/getUserInfoByName", controller.GetUserInfoByNameController)
		// TODO:测试使用
		apiUser.GET("/test", test)
	}
	apiGroup := router.Group("/api/group")
	apiGroup.Use(middleware.JWT())
	{
		// 创建群聊
		apiGroup.POST("/create", controller.CreateGroupController)
		// 删除群聊
		apiGroup.DELETE("/:id", service.DeleteGroupService)
		// 加入群聊
		apiGroup.POST("/join", service.JoinGroupService)
		// 退出群聊
		apiGroup.POST("/quit", service.QuitGroupService)
		// 获取群聊成员列表
		apiGroup.GET("/getGroupMemberList", service.GetGroupMemberListService)
		// 授予管理员权限
		apiGroup.POST("/grantAdministrator", service.GrantAdministratorService)
		// 移除管理员权限
		apiGroup.POST("/removeAdministrator", service.RemoveAdministratorService)
	}

	apiContact := router.Group("/api/contact")
	apiContact.Use(middleware.JWT())
	{
		// 获取联系人列表
		apiContact.GET("getContactList", controller.GetContactListController)
		// 获取新的联系人列表
		apiContact.GET("getNewContactList", controller.GetNewContactListController)
		// 获取联系人详情
		apiContact.GET("getMessageList", service.GetContactDetailService)
		// 获取新的消息列表
		apiContact.GET("getNewMsgList", controller.GetNewMsgListController)
		// 获取批量用户信息
		apiContact.POST("userInfo/batch", controller.GetUserInfoBatchController)
	}

	apiMsg := router.Group("/api/chat")
	apiMsg.Use(middleware.JWT())
	{
		// 发送消息
		apiMsg.POST("msg", controller.SendMessageController)
	}

	apiFile := router.Group("/api/file")
	apiFile.Use(middleware.JWT())
	{
		// 上传文件
		apiFile.GET("getPreSigned", service.GetPreSigned)
	}

	err := router.Run(":5000")
	if err != nil {
		return
	}
}

// TODO:测试使用
func test(c *gin.Context) {
	msg := new(global.Msg)
	msg.Uid = 2
	websocketService.Send(msg.Uid, []byte("{\"type\":4}"))
	resp.SuccessResponse(c, nil)
}
