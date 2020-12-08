package service

type CommonReturn struct {
	Code   int
	Reason string
}
type SecKillProductInfo struct {
	ProductID int
	Start     bool
	End       bool
	Status    int
	Total     int
	Left      int
}
type ReadSecProRsp struct {
	Ret  *CommonReturn
	Info *SecKillProductInfo
}
