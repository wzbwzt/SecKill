package conf

import (
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
)

var (
	secKillConfig *SysConfig
	redisPool     *redis.Pool
)

type SysConfig struct {
	redisConf *RedisConfig
	etcdConf  *EtcdConfig
}
type RedisConfig struct {
	redisAddr   string
	maxIdle     int
	maxActive   int
	idleTimeOut int
}

type EtcdConfig struct {
	etcdAddr string
	timeOut  int
}

func init() {
	//初始化配置文件
	initConfig()

	var err error
	//加载redis配置
	err = initRedis()
	if err != nil {
		logs.Debug("init redis failed,err:%s", err)
	}
	//加载etcd配置
	initEtcd()

}

func initConfig() {
	secKillConfig = &SysConfig{
		etcdConf: &EtcdConfig{
			etcdAddr: beego.AppConfig.String("etcdAddr"),
			timeOut:  beego.AppConfig.DefaultInt("etcdTimeOut", 10),
		},
		redisConf: &RedisConfig{
			redisAddr:   beego.AppConfig.String("redisAddr"),
			maxIdle:     beego.AppConfig.DefaultInt("maxIdle", 10),
			maxActive:   beego.AppConfig.DefaultInt("maxActive", 10),
			idleTimeOut: beego.AppConfig.DefaultInt("idleTimeOut", 10),
		},
	}
}

func initRedis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     secKillConfig.redisConf.maxIdle,
		MaxActive:   secKillConfig.redisConf.maxActive,
		IdleTimeout: time.Duration(secKillConfig.redisConf.idleTimeOut) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secKillConfig.redisConf.redisAddr)
		},
	}

	//测试连接池
	conn := redisPool.Get()
	_, err = conn.Do("ping")
	if err != nil {
		return
	}
	return
}
func initEtcd() {

}
