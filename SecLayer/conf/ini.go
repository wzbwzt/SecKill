package conf

import (
	"sync"

	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
)

var (
	SecLayerSysConfig *SecLayerConf
)

type RedisConf struct {
	RedisAddr        string
	RedisPwd         string
	RedisMaxIdle     int
	RedisMaxActive   int
	RedisIdleTimeout int
	RedisQueueName   string
}

type EtcdConf struct {
	EtcdAddr          string
	Timeout           int
	EtcdSecKeyPrefix  string
	EtcdSecProductKey string
}

type SecLimit struct {
	count   int   //每秒访问数量
	curTime int64 //访问的时间(精确到秒)
}

type SecProductInfoConf = struct {
	ProductId         int
	StartTime         int64
	EndTime           int64
	Status            int
	Total             int
	Left              int
	OnePersonBuyLimit int
	BuyRate           float64
	SoldMaxLimit      int       //每秒组多能卖多少个
	secLimit          *SecLimit //限速控制
}

type SecLayerConf struct {
	Proxy2LayerRedis RedisConf
	Layer2ProxyRedis RedisConf
	EtcdConfig       EtcdConf
	LogPath          string
	LogLevel         string

	WriteGoroutineNum      int
	ReadGoroutineNum       int
	HandleUserGoroutineNum int
	Read2HandleChanSize    int
	Handle2WriteChanSize   int
	MaxRequestWaitTimeout  int

	SendToWriteChanTimeout  int
	SendToHandleChanTimeout int

	SecProductInfoMap        map[int]*SecProductInfoConf
	SecProductInfoUpdateLock sync.RWMutex
	TokenPassword            string
}

func Init(typeName, path string) (err error) {
	configer, err := config.NewConfig(typeName, path)
	if err != nil {
		logs.Error("new config failed ,err:", err.Error())
		return
	}

	SecLayerSysConfig = &SecLayerConf{
		Proxy2LayerRedis: RedisConf{
			RedisAddr:        configer.String("redis::redis_proxy2layer_addr"),
			RedisPwd:         configer.String("redis::redis_proxy2layer_pwd"),
			RedisMaxIdle:     configer.DefaultInt("redis::redis_proxy2layer_idle", 10),
			RedisMaxActive:   configer.DefaultInt("redis::redis_proxy2layer_active", 10),
			RedisIdleTimeout: configer.DefaultInt("redis::redis_proxy2layer_timeout", 200),
			RedisQueueName:   configer.DefaultString("redis::redis_proxy2layer_queue_name", "sec_queue"),
		},
		Layer2ProxyRedis: RedisConf{
			RedisAddr:        configer.String("redis::redis_layer2proxy_addr"),
			RedisPwd:         configer.String("redis::redis_layer2proxy_pwd"),
			RedisMaxIdle:     configer.DefaultInt("redis::redis_layer2proxy_idle", 10),
			RedisMaxActive:   configer.DefaultInt("redis::redis_layer2proxy_active", 10),
			RedisIdleTimeout: configer.DefaultInt("redis::redis_layer2proxy_timeout", 200),
			RedisQueueName:   configer.DefaultString("redis::redis_layer2proxy_queue_name", "sec_queue"),
		},
		EtcdConfig: EtcdConf{
			EtcdAddr:          configer.DefaultString("etcd::server_addr", "127.0.0.1:2379"),
			EtcdSecKeyPrefix:  configer.DefaultString("etcd::etcd_sec_key_prefix", "/joelWu/backend/seckill"),
			Timeout:           configer.DefaultInt("etcd::etcd_timeout", 5),
			EtcdSecProductKey: configer.DefaultString("etcd::etcd_product_key", "product"),
		},
		LogPath:                configer.DefaultString("logs::log_path", "./logs/seclayer.log"),
		LogLevel:               configer.DefaultString("logs::log_level", "debug"),
		ReadGoroutineNum:       configer.DefaultInt("redis::read_goroutine_num", 10),
		WriteGoroutineNum:      configer.DefaultInt("redis::write_goroutine_numetcd_timeout", 10),
		HandleUserGoroutineNum: configer.DefaultInt("redis::handle_user_goroutine_numetcd_timeout", 10),
	}
	return
}
