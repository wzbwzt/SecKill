package routers

import (
	"SecProxy/controller"
	"SecProxy/controllers"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/seckillList", &controller.SecKillController{}, "*:SecKillProdList")
	beego.Router("/secinfo", &controller.SecKillController{}, "*:SecProdInfo")
	beego.Router("/seckill", &controller.SecKillController{}, "post:SecKill")
}
