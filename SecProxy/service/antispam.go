package service

import (
	"SecProxy/conf"
	"SecProxy/parameter"
	"sync"

	"github.com/astaxie/beego/logs"
)

//检测每秒访问速度，对于机器访问的禁止访问
var (
	secLimitMgr *SecLimitMgr = &SecLimitMgr{UserSecLimit: make(map[int64]*SecLimit)}
)

type SecLimitMgr struct {
	UserSecLimit map[int64]*SecLimit  //用户维度的访问频率限制
	IPSecLimit   map[string]*SecLimit //IP维度的访问频率限制
	Lock         sync.Mutex
}

type SecLimit struct {
	count   int   //每秒访问数量
	curTime int64 //访问的时间(精确到秒)
}

//Count 更新每秒访问数量
func (s *SecLimit) Count(nowTime int64) (newCount int) {
	if s.curTime == nowTime {
		return s.count + 1
	}
	s.count = 1
	s.curTime = nowTime
	return 1
}

func antispam(req *parameter.SecKillReq) (err error) {
	secLimitMgr.Lock.Lock()
	defer secLimitMgr.Lock.Unlock()

	_val_user_seclimit, ok := secLimitMgr.UserSecLimit[req.UserID]
	if !ok {
		_val_user_seclimit = &SecLimit{}
		secLimitMgr.UserSecLimit[req.UserID] = _val_user_seclimit
	}
	newcount := _val_user_seclimit.Count(req.AccessTime.Unix())
	if newcount > conf.SecKillConfig.MaxSecAccessLimit {
		logs.Warn("user:%d is reject by out of SecLimit", req.UserID)
		return New(ErrUserServiceBusy, "非法用户访问")
	}

	_val_ip_seclimit, ok := secLimitMgr.IPSecLimit[req.ClientAddr]
	if !ok {
		_val_ip_seclimit = &SecLimit{}
		secLimitMgr.IPSecLimit[req.ClientAddr] = _val_ip_seclimit
	}
	ip_newcount := _val_ip_seclimit.Count(req.AccessTime.Unix())
	if ip_newcount > conf.SecKillConfig.IPSecAccessLimit {
		logs.Warn("ip:%s is reject by out of secLimit", req.ClientAddr)
		return New(ErrUserServiceBusy, "非法ip访问")
	}

	return
}
