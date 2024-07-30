package main

import (
	_ "DiTing-Go/event/listener"
	"DiTing-Go/routes"
)

// swagger 中添加header.Authorization:token 校验 token
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	//global.InitDB()
	//dal.DB.AutoMigrate(&model.User{}) // sql文件中没有user表，这里手动导入一下
	routes.InitRouter()

}
