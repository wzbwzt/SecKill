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
	ProductID     int64
	Source        string //来源
	AuthCode      string
	SecTime       string //抢购时间
	Nance         string
	UserID        int64  //用于校验用户是否处于登录状态
	UserAuthSign  string //用于校验用户是否处于登录状态
	AccessTime    time.Time
	ClientAddr    string //获取请求的IP地址，对恶意访问的地址做限制
	ClientRefence string //获取访问的来源，对于非正规流程过来的请求(如非抢购页面过来的请求)，做限制
}
