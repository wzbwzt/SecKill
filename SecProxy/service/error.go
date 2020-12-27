package service

const (
	ErrCodeSuccess         = 1000
	ErrInvalidParam        = 1001
	ErrNotFoundSource      = 1002
	ErrUserCheckAuthFailed = 1003
	ErrUserServiceBusy     = 1004
	ErrActiveNotStart      = 1005
	ErrActiveAlreadyEnd    = 1006
	ErrActiveSaleOut       = 1007
	ErrProcessTimeout      = 1008
	ErrClientClosed        = 1009
)

type MyErr struct {
	Code   int
	Reason string
}

func (m MyErr) Error() string {
	return m.Reason
}

func New(code int, reason string) error {
	return MyErr{
		Code:   code,
		Reason: reason,
	}
}
