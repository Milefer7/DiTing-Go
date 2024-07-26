package controller

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/global"
	pkgReq "DiTing-Go/pkg/domain/vo/req"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

// GetUserInfoBatchController 获取批量用户信息
//
//	@Summary	获取批量用户信息
//	@Produce	json
//	@Param		uidList	query		[]int64			true	"用户ID列表"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/userInfo/batch [post]
func GetUserInfoBatchController(c *gin.Context) {
	getUserInfoBatchReq := req.GetUserInfoBatchReq{}
	if err := c.ShouldBind(&getUserInfoBatchReq); err != nil {
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetUserInfoBatchService(getUserInfoBatchReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
	return
}

// GetContactListController 获取联系人列表
//
//	@Summary	获取联系人列表
//	@Produce	json
//	@Param		uid	query		int64				true	"用户ID"
//	@Param		cursor	query		string				false	"游标，用于分页查询"
//	@Param		pageSize	query		int				false	"每页数量，默认为20"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/getContactList [get]
func GetContactListController(c *gin.Context) {
	uid := c.GetInt64("uid")
	// 游标翻页
	// 默认值
	var cursor *string = nil
	var pageSize int = 20
	pageRequest := pkgReq.PageReq{
		Cursor:   cursor,
		PageSize: pageSize,
	}
	if err := c.ShouldBindQuery(&pageRequest); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetContactListService(uid, pageRequest)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// GetNewContactListController 获取新的联系人列表
//
//	@Summary	获取新的联系人列表
//	@Produce	json
//	@Param		uid	query		int64				true	"用户ID"
//	@Param		timestamp	query		int64				true	"时间戳，用于获取此时间之后的新联系人"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/getNewContactList [get]
func GetNewContactListController(c *gin.Context) {
	uid := c.GetInt64("uid")

	getNewContentListReq := req.GetNewContentListReq{}
	if err := c.ShouldBindQuery(&getNewContentListReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetNewContactListService(uid, getNewContentListReq.Timestamp)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// GetNewMsgListController 获取新的消息列表
//
//	@Summary	获取新的消息列表
//	@Produce	json
//	@Param		msgId	query		int64				true	"消息ID"
//	@Param		roomId	query		int64				true	"房间ID"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/contact/getNewMsgList [get]
func GetNewMsgListController(c *gin.Context) {

	getNewMsgListReq := req.GetNewMsgListReq{}
	if err := c.ShouldBindQuery(&getNewMsgListReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.GetNewMsgService(getNewMsgListReq.MsgId, getNewMsgListReq.RoomId)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}

// CreateGroupController 创建群聊
//
//	@Summary	创建群聊
//	@Produce	json
//	@Param		uid	query		int64				true	"用户ID"
//	@Param		uidList	query		[]int64			true	"群聊成员ID列表"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/group/create [post]
func CreateGroupController(c *gin.Context) {
	uid := c.GetInt64("uid")
	creatGroupReq := req.CreateGroupReq{}
	if err := c.ShouldBind(&creatGroupReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		global.Logger.Errorf("参数错误: %v", err)
		c.Abort()
		return
	}
	response, err := service.CreateGroupService(uid, creatGroupReq.UidList)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
