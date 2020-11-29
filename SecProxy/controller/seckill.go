package controller

import "github.com/astaxie/beego"

type SecKillController struct {
	beego.Controller
}

func (s *SecKillController) SecKill() {
	s.Data["json"] = "hello beego"
	s.ServeJSON()

}
func (s *SecKillController) SecInfo() {
	s.Data["json"] = "hello beego"
	s.ServeJSON()
}
