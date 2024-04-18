package controller

import (
	"DiTing-Go/domain/vo/req"
	"DiTing-Go/pkg/e"
	"DiTing-Go/pkg/resp"
	"DiTing-Go/service"
	"github.com/gin-gonic/gin"
)

// RegisterController 用户注册
//
//	@Summary	用户注册
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/register [post]
func RegisterController(c *gin.Context) {
	userReq := req.UserRegisterReq{}
	if err := c.ShouldBind(&userReq); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response := service.RegisterService(userReq)
	if response.Code == e.ERROR {
		c.Abort()
	}
	resp.ReturnResponse(c, response)
}

// LoginController 用户登录
//
//	@Summary	用户登录
//	@Produce	json
//	@Param		name		body		string				true	"用户名"
//	@Param		password	body		string				true	"密码"
//	@Success	200			{object}	resp.ResponseData	"成功"
//	@Failure	500			{object}	resp.ResponseData	"内部错误"
//	@Router		/api/public/login [post]
func LoginController(c *gin.Context) {
	userLogin := req.UserLoginReq{}
	if err := c.ShouldBind(&userLogin); err != nil { //ShouldBind()会自动推导
		resp.ErrorResponse(c, "参数错误")
		c.Abort()
		return
	}
	response := service.LoginService(userLogin)
	if response.Code == e.ERROR {
		c.Abort()
	}
	resp.ReturnResponse(c, response)
}