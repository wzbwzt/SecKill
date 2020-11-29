package routers

import (
	"SecProxy/controller"
	"SecProxy/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/seckill", &controller.SecKillController{}, "*:SecKill")
	beego.Router("/secinfo", &controller.SecKillController{}, "*:SecInfo")
}
