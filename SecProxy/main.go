package main

import (
	_ "SecProxy/conf"
	_ "SecProxy/routers"

	"github.com/astaxie/beego"
)

func main() {
	beego.Run()

}
