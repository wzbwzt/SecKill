package parameter

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
