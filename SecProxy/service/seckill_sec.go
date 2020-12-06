package service

import (
	"SecProxy/conf"

	"github.com/micro/go-micro/v2/util/log"
)

func ReadSecKilProInfo(id int) (out *ReadSecProRsp, err error) {
	if id == 0 {
		out.Ret = &CommonReturn{
			Code:   ErrInvalidParam,
			Reason: "ID must request",
		}
		return
	}

	log.Debug("data:1231313s")
	data := conf.MapSecKillProducts
	log.Debug("data:", data)
	if info, ok := data[id]; ok {
		out.Info = &SecKillProductInfo{
			ProductID: info.ProductID,
			StartTime: info.StartTime,
			EndTime:   info.EndTime,
			Total:     info.Total,
			Left:      info.Left,
			Status:    info.Status,
		}
		return
	}

	out.Ret = &CommonReturn{
		Code:   ErrNotFoundSource,
		Reason: "product has not exist",
	}
	return
}
