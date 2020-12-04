package conf

import (
	"encoding/json"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
	etcd "go.etcd.io/etcd/clientv3"
)

var (
	secKillConfig *SysConfig
	redisPool     *redis.Pool
	etcdClient    *etcd.Client
)

type SysConfig struct {
	redisConf *RedisConfig
	etcdConf  *EtcdConfig
	logConf   *LogsConfig
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

type LogsConfig struct {
	logPath  string
	logLevel string
}

func init() {
	//初始化配置文件
	initConfig()

	var err error
	//加载redis配置
	err = initRedis()
	if err != nil {
		logs.Debug("init redis failed,err:%s", err)
		panic(err)
	}
	//加载etcd配置
	err = initEtcd()
	if err != nil {
		logs.Debug("init etcd failed,err:%s", err)
		panic(err)
	}

	//加载日志配置文件
	err = initLogs()
	if err != nil {
		logs.Debug("init logsconfig failed,err:%s", err)
		panic(err)
	}
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
		logConf: &LogsConfig{
			logPath:  beego.AppConfig.String("logPath"),
			logLevel: beego.AppConfig.String("logLevel"),
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

func initEtcd() (err error) {
	etcdClient, err = etcd.New(etcd.Config{
		Endpoints:   []string{},
		DialTimeout: time.Duration(secKillConfig.etcdConf.timeOut) * time.Millisecond,
	})

	if err != nil {
		return
	}
	return
}

func initLogs() (err error) {
	config := make(map[string]interface{})
	config["filename"] = secKillConfig.logConf.logPath
	config["level"] = transferLogLevel(secKillConfig.logConf.logLevel)
	var jstr []byte
	jstr, err = json.Marshal(config)
	if err != nil {
		return
	}
	err = logs.SetLogger(logs.AdapterFile, string(jstr))
	if err != nil {
		return
	}
	return
}

func transferLogLevel(level string) int {
	switch level {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	default:
		return logs.LevelDebug
	}
}
