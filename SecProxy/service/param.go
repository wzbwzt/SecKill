package service

type CommonReturn struct {
	Code   int
	Reason string
}
type SecKillProductInfo struct {
	ProductID int
	StartTime int64
	EndTime   int64
	Status    int
	Total     int
	Left      int
}
type ReadSecProRsp struct {
	Ret  *CommonReturn
	Info *SecKillProductInfo
}
