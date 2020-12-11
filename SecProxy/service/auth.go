package service

import (
	"SecProxy/conf"
	"crypto/md5"
	"errors"
	"fmt"
)

//密钥验证
func userCheck(userID int64, reqAuthSign string) (err error) {
	source := fmt.Sprintf("%s-%v", conf.SecKillConfig.SecretKey, userID)
	authSign := fmt.Sprintf("%x", md5.Sum([]byte(source)))
	if string(authSign) != reqAuthSign {
		return errors.New("involid user cookie auth")
	}
	return

}
