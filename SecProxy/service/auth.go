package service

import (
	"SecProxy/conf"
	"SecProxy/parameter"
	"crypto/md5"
	"errors"
	"fmt"

	"github.com/astaxie/beego/logs"
)

//密钥验证;白名单验证
func userCheck(req *parameter.SecKillReq) (err error) {
	source := fmt.Sprintf("%s-%v", conf.SecKillConfig.SecretKey, req.UserID)
	authSign := fmt.Sprintf("%x", md5.Sum([]byte(source)))
	if string(authSign) != req.UserAuthSign {
		return errors.New("involid user cookie auth")
	}

	refence := false
	if len(conf.SecKillConfig.RefenceWhiteList) != 0 {
		for _, v := range conf.SecKillConfig.RefenceWhiteList {
			if v == req.ClientRefence {
				refence = true
				break
			}
		}
	}
	if !refence {
		logs.Warn("user:%d is reject by refene[%s]", req.UserID, req.ClientRefence)
		return New(ErrInvalidParam, "异常访问来源")
	}
	return

}
