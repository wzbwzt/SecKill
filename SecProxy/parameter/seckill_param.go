package parameter

type ReadSecKProdInfoRsp struct {
	ProductID int   `json:"ProductID"`
	StartTime int64 `json:"StartTime"`
	EndTime   int64 `json:"EndTime  "`
	Status    int   `json:"Status   "`
	Total     int   `json:"Total    "`
	Left      int   `json:"Left     "`
}
