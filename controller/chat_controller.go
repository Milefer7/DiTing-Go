package controller

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/pkg/domain/vo/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

// SendMessageController 发送消息
//
//	@Summary	发送消息
//	@Produce	json
//	@Param		messageReq	body		req.MessageReq	true	"消息请求体"
//	@Success	200	{object}	resp.ResponseData	"成功"
//	@Failure	500	{object}	resp.ResponseData	"内部错误"
//	@Router		/api/chat/msg [post]
func SendMessageController(c *gin.Context) {
	uid := c.GetInt64("uid")
	messageReq := req.MessageReq{}
	if err := c.ShouldBind(&messageReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response, err := service.SendTextMsgService(uid, messageReq)
	if err != nil {
		c.Abort()
		resp.ReturnErrorResponse(c, response)
		return
	}
	resp.ReturnSuccessResponse(c, response)
}
