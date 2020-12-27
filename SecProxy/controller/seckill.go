package controller

import (
	"SecProxy/parameter"
	"SecProxy/service"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
)

type SecKillController struct {
	beego.Controller
}

func (s *SecKillController) SecKillProdList() {
	rsp := parameter.Response{}
	defer func() {
		s.Data["json"] = rsp
		s.ServeJSON()
	}()

	rsp.Code = 0
	rsp.Msg = "succes"
	return
}

//SecProdInfo 商品详情
func (s *SecKillController) SecProdInfo() {
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

func (s *SecKillController) SecKill() {
	rsp := &parameter.Response{}
	defer func() {
		s.Data["json"] = rsp
		s.ServeJSON()
	}()

	userID, err := strconv.ParseInt(s.Ctx.GetCookie("userID"), 10, 64)
	if err != nil {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = err.Error()
		return
	}
	reqAddr := s.Ctx.Request.RemoteAddr
	if len(strings.Split(reqAddr, ":")) == 0 {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = "无效的请求地址"
		return
	}
	reqIP := strings.Split(reqAddr, ":")[0]

	prodID, err := s.GetInt("prodId")
	if err != nil {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = err.Error()
		return
	}

	req := &parameter.SecKillReq{
		ClientAddr:    reqIP,
		UserID:        userID,
		UserAuthSign:  s.Ctx.GetCookie("userAuthSign"),
		ProductID:     int64(prodID),
		Source:        s.GetString("source"),
		AuthCode:      s.GetString("authcode"),
		SecTime:       s.GetString("time"),
		Nance:         s.GetString("nance"),
		ClientRefence: s.Ctx.Request.Referer(),
		AccessTime:    time.Now(),
	}

	out, err := service.SecKill(req)
	if err != nil {
		rsp.Code = service.ErrInvalidParam
		rsp.Msg = err.Error()
		return
	}
	if out.Ret.Code != service.ErrCodeSuccess {
		rsp.Code = out.Ret.Code
		rsp.Msg = out.Ret.Reason
		return
	}

	rsp.Code = service.ErrCodeSuccess
	rsp.Msg = "抢购成功"
	return
}
