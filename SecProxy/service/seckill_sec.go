package service

import (
	"SecProxy/conf"
	"SecProxy/parameter"
	"time"
)

func ReadSecKilProInfo(id int) (out *ReadSecProRsp, err error) {
	if id == 0 {
		out = &ReadSecProRsp{
			Ret: &CommonReturn{
				Code:   ErrInvalidParam,
				Reason: "ID must request",
			},
		}
		return
	}

	data := conf.MapSecKillProducts

	if info, ok := data[id]; ok {
		now := time.Now().Unix()
		var start, end bool
		var status int
		if now-info.StartTime < 0 {
			out = &ReadSecProRsp{
				Ret: &CommonReturn{
					Code:   ErrActiveNotStart,
					Reason: "active has not start",
				},
			}
			start = false
			end = false
			status = parameter.OnSale
		}

		if now-info.StartTime > 0 && now < info.EndTime {
			start = true
			end = false
			status = parameter.OnSale
		}

		if info.Status == parameter.ForceSaleOut || info.Status == parameter.HasSaleOut {
			start = false
			end = true
			status = parameter.HasSaleOut
		}

		if now-info.EndTime > 0 {
			out = &ReadSecProRsp{
				Ret: &CommonReturn{
					Code:   ErrActiveAlreadyEnd,
					Reason: "active has already end",
				},
			}
			start = false
			end = true
			status = parameter.HasSaleOut
		}

		out = &ReadSecProRsp{
			Info: &SecKillProductInfo{
				ProductID: info.ProductID,
				Total:     info.Total,
				Left:      info.Left,
				Status:    status,
				Start:     start,
				End:       end,
			},
		}
		return
	}

	out = &ReadSecProRsp{
		Ret: &CommonReturn{
			Code:   ErrNotFoundSource,
			Reason: "product has not exist",
		},
	}
	return
}
