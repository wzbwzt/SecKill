package routers

import (
	"SecProxy/controller"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/seckillList", &controller.SecKillController{}, "*:SecKillProdList")
	beego.Router("/secinfo", &controller.SecKillController{}, "*:SecProdInfo")
	beego.Router("/seckill", &controller.SecKillController{}, "post:SecKill")
}
