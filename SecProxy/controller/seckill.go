package controller

import (
	"SecProxy/parameter"
	"SecProxy/service"

	"github.com/astaxie/beego"
)

type SecKillController struct {
	beego.Controller
}

func (s *SecKillController) SecKill() {
	rsp := parameter.Response{}
	defer func() {
		s.Data["json"] = rsp
		s.ServeJSON()
	}()

	rsp.Code = 0
	rsp.Msg = "succes"

}
func (s *SecKillController) SecInfo() {
	rsp := &parameter.Response{}
	defer func() {
		s.Data["json"] = rsp
		s.ServeJSON()
	}()
	prodID, err := s.GetInt("prodId")
	if err != nil {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = "invoild body"
		return
	}

	grsp, err := service.ReadSecKilProInfo(prodID)

	if err != nil {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = err.Error()
		return
	}
	if grsp.Ret != nil && grsp.Ret.Code != service.ErrCodeSuccess {
		rsp.Code = grsp.Ret.Code
		rsp.Msg = grsp.Ret.Reason
		return
	}
	//不对外展示时间，每个客户端的时间可能不一致，无法做到统一处理；所以得以服务器时间为准，只返回活动
	//开始与否以及状态
	data := parameter.ReadSecKProdInfoRsp{
		ProductID: grsp.Info.ProductID,
		Total:     grsp.Info.Total,
		Left:      grsp.Info.Left,
		Status:    grsp.Info.Status,
		Start:     grsp.Info.Start,
		End:       grsp.Info.End,
	}
	rsp.Data = data
	rsp.Code = service.ErrCodeSuccess
	rsp.Msg = "success"

	return
}
