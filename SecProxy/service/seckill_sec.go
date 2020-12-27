package service

import (
	"SecProxy/conf"
	"SecProxy/parameter"
	"sync"
	"time"
)

var (
	RWlock sync.RWMutex
)

//ReadSecKilProInfo 获取秒杀商品详情
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
			return
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
					Reason: "this product has already end",
				},
			}
			return
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

func SecKill(req *parameter.SecKillReq) (out *ReadSecProRsp, err error) {
	RWlock.RLock()
	defer RWlock.RUnlock()

	//校验用户是否登录
	err = userCheck(req)
	if err != nil {
		if myerr, ok := err.(MyErr); ok {
			out.Ret = &CommonReturn{
				Code:   myerr.Code,
				Reason: myerr.Reason,
			}
			return
		}
		return
	}

	//用户id和ip访问频率检测更新
	err = antispam(req)
	if err != nil {
		if myerr, ok := err.(MyErr); ok {
			out.Ret = &CommonReturn{
				Code:   myerr.Code,
				Reason: myerr.Reason,
			}
			return
		}
		return
	}

	res, err := ReadSecKilProInfo(int(req.ProductID))
	if err != nil {
		return
	}
	if res.Ret.Code != ErrCodeSuccess {
		out.Ret = &CommonReturn{
			Code:   res.Ret.Code,
			Reason: res.Ret.Reason,
		}
		return
	}

	SecReqChan <- req

	return
}
