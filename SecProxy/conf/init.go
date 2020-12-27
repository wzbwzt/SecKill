package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	etcd "go.etcd.io/etcd/clientv3"
)

var (
	SecKillConfig        *SysConfig
	BlackRedisPool       *redis.Pool
	Proxy2LayerRedisPool *redis.Pool
	etcdClient           *etcd.Client
	MapSecKillProducts   = make(map[int]*SecProductInfoConf)
)

type SysConfig struct {
	redisBlackConf        *RedisConfig
	redisProxy2LayerConf  *RedisConfig
	etcdConf              *EtcdConfig
	logConf               *LogsConfig
	RwLock                sync.RWMutex
	SecretKey             string
	RefenceWhiteList      []string
	IPSecAccessLimit      int
	MaxSecAccessLimit     int
	IPBlcakMap            map[string]bool
	UserIDBlcakMap        map[int64]bool
	SyncRwLock            sync.RWMutex
	WriteProxy2LayerGoNum int
	ReadLayer2ProxyGoNum  int
	SecReqChanSize        int
}
type RedisConfig struct {
	redisAddr   string
	maxIdle     int
	maxActive   int
	idleTimeOut int
}

type EtcdConfig struct {
	etcdAddr       string
	timeOut        int
	etcdSecKPrefix string
	etcdProductKey string
}

type LogsConfig struct {
	logPath  string
	logLevel string
}

type SecProductInfoConf struct {
	ProductID int
	StartTime int64
	EndTime   int64
	Status    int
	Total     int
	Left      int
}

func init() {
	//初始化配置文件
	initConfig()

	var err error
	//加载黑名单redis配置
	err = initRedisForBlackList()
	if err != nil {
		logs.Error("init redis failed,err:%s", err)
		panic(err)
	}
	//加载接入层到逻辑层redis配置
	err = initRedisForProxy2Layer()
	if err != nil {
		logs.Error("init redis failed,err:%s", err)
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
		logs.Error("init logsconfig failed,err:%s", err)
		panic(err)
	}

	//从etcd中加载秒杀文件
	err = loadSecConfig()
	if err != nil {
		logs.Error("load seckill config failed,err:%s", err)
		panic(err)
	}

	// 监听秒杀配置是否改变
	initSecProductWatcher()

}

func initConfig() {
	refenceWhiteList := strings.Split(beego.AppConfig.String("refenceWhiteList"), ",")
	SecKillConfig = &SysConfig{
		etcdConf: &EtcdConfig{
			etcdAddr:       beego.AppConfig.String("etcdAddr"),
			timeOut:        beego.AppConfig.DefaultInt("etcdTimeOut", 10),
			etcdSecKPrefix: beego.AppConfig.String("etcdSecKeyPrefix"),
			etcdProductKey: beego.AppConfig.String("etcdProductKey"),
		},
		redisBlackConf: &RedisConfig{
			redisAddr:   beego.AppConfig.String("redisBlackAddr"),
			maxIdle:     beego.AppConfig.DefaultInt("blackMaxIdle", 10),
			maxActive:   beego.AppConfig.DefaultInt("blackMaxActive", 10),
			idleTimeOut: beego.AppConfig.DefaultInt("blackIdleTimeOut", 10),
		},
		redisProxy2LayerConf: &RedisConfig{
			redisAddr:   beego.AppConfig.String("proxy2LayerRedisAddr"),
			maxIdle:     beego.AppConfig.DefaultInt("proxy2LayerMaxIdle", 10),
			maxActive:   beego.AppConfig.DefaultInt("proxy2LayerMaxActive", 10),
			idleTimeOut: beego.AppConfig.DefaultInt("proxy2LayerIdleTimeOut", 10),
		},
		logConf: &LogsConfig{
			logPath:  beego.AppConfig.String("logPath"),
			logLevel: beego.AppConfig.String("logLevel"),
		},
		SecretKey:             beego.AppConfig.String("secretKey"),
		RefenceWhiteList:      refenceWhiteList,
		IPSecAccessLimit:      beego.AppConfig.DefaultInt("ipSecAccessLimit", 10),
		MaxSecAccessLimit:     beego.AppConfig.DefaultInt("maxSecAccessLimit", 10),
		WriteProxy2LayerGoNum: beego.AppConfig.DefaultInt("writeProxy2LayerGoroutineNum", 10),
		ReadLayer2ProxyGoNum:  beego.AppConfig.DefaultInt("readLayer2ProxyGoroutineNum", 10),
		SecReqChanSize:        beego.AppConfig.DefaultInt("secReqChanSize", 100),
	}
}

//加载黑名单的redis配置
func initRedisForBlackList() (err error) {
	BlackRedisPool = &redis.Pool{
		MaxIdle:     SecKillConfig.redisBlackConf.maxIdle,
		MaxActive:   SecKillConfig.redisBlackConf.maxActive,
		IdleTimeout: time.Duration(SecKillConfig.redisBlackConf.idleTimeOut) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", SecKillConfig.redisBlackConf.redisAddr)
		},
	}

	//测试连接池
	conn := BlackRedisPool.Get()
	_, err = conn.Do("ping")
	if err != nil {
		return
	}
	return
}

//加载接入层到逻辑层redis配置
func initRedisForProxy2Layer() (err error) {
	Proxy2LayerRedisPool = &redis.Pool{
		MaxIdle:     SecKillConfig.redisProxy2LayerConf.maxIdle,
		MaxActive:   SecKillConfig.redisProxy2LayerConf.maxActive,
		IdleTimeout: time.Duration(SecKillConfig.redisProxy2LayerConf.idleTimeOut) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", SecKillConfig.redisProxy2LayerConf.redisAddr)
		},
	}

	//测试连接池
	conn := Proxy2LayerRedisPool.Get()
	_, err = conn.Do("ping")
	if err != nil {
		return
	}
	return
}
func initEtcd() (err error) {
	etcdClient, err = etcd.New(etcd.Config{
		Endpoints:   []string{SecKillConfig.etcdConf.etcdAddr},
		DialTimeout: time.Duration(SecKillConfig.etcdConf.timeOut) * time.Millisecond,
	})

	if err != nil {
		return
	}
	return
}

func initLogs() (err error) {
	config := make(map[string]interface{})
	config["filename"] = SecKillConfig.logConf.logPath
	config["level"] = transferLogLevel(SecKillConfig.logConf.logLevel)
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

//从etd加载秒杀配置
func loadSecConfig() (err error) {
	key := fmt.Sprintf("%s/%s", SecKillConfig.etcdConf.etcdSecKPrefix, SecKillConfig.etcdConf.etcdProductKey)
	rsp, err := etcdClient.Get(context.Background(), key)
	if err != nil {
		return
	}

	var secProductInfo []SecProductInfoConf
	for k, v := range rsp.Kvs {
		logs.Debug("key:%s value:%s", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			return
		}
		logs.Debug("secInfo conf is %v", secProductInfo)
	}

	SecKillConfig.RwLock.Lock()
	for _, v := range secProductInfo {
		_, ok := MapSecKillProducts[v.ProductID]
		if ok {
			continue
		}
		tmp := v
		MapSecKillProducts[v.ProductID] = &tmp
	}
	SecKillConfig.RwLock.Unlock()
	return
}

//监听秒杀配置是否改变
func initSecProductWatcher() {
	key := fmt.Sprintf("%s/%s", SecKillConfig.etcdConf.etcdSecKPrefix, SecKillConfig.etcdConf.etcdProductKey)
	go watchSecProductKey(key)
}

func watchSecProductKey(key string) {
	logs.Debug("begin watch key: %s", key)

	for {
		watchChan := etcdClient.Watch(context.Background(), key)
		var secProductInfo []SecProductInfoConf
		var getConfSucc = true

		for wresp := range watchChan {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s]'s config deleted", key)
					continue
				}
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err := json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key[%s],Unmarshal[%s] failed,err: %v", ev.Kv.Key, ev.Kv.Value, err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd,%s %q: %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}

			if getConfSucc {
				logs.Debug("get config from etcd success, %v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}
	}
}

func updateSecProductInfo(confs []SecProductInfoConf) {
	//每个加锁效率会是问题，可以先放入到一个临时变量中，再加锁赋值
	tmp := make(map[int]*SecProductInfoConf)
	for _, v := range confs {
		ttmp := v
		tmp[v.ProductID] = &ttmp
	}

	SecKillConfig.RwLock.Lock()
	MapSecKillProducts = tmp
	SecKillConfig.RwLock.Unlock()
}
