package parameter

import "time"

const (
	OnSale = iota
	HasSaleOut
	ForceSaleOut
)

type ReadSecKProdInfoRsp struct {
	ProductID int  `json:"productID"`
	Start     bool `json:"start"`
	End       bool `json:"end"`
	Status    int  `json:"status"`
	Total     int  `json:"total"`
	Left      int  `json:"left"`
}

type SecKillReq struct {
	ProductID    int64
	Source       string //来源
	AuthCode     string
	SecTime      string //抢购时间
	Nance        string
	UserID       string //用于校验用户是否处于登录状态
	UserAuthSign string //用于校验用户是否处于登录状态
	AccessTime   time.Time
}
