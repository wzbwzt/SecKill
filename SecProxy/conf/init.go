package conf

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	etcd "go.etcd.io/etcd/clientv3"
)

var (
	SecKillConfig      *SysConfig
	redisPool          *redis.Pool
	etcdClient         *etcd.Client
	MapSecKillProducts = make(map[int]*SecProductInfoConf)
)

type SysConfig struct {
	redisConf *RedisConfig
	etcdConf  *EtcdConfig
	logConf   *LogsConfig
	RwLock    sync.RWMutex
	SecretKey string
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
	//加载redis配置
	err = initRedis()
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
	SecKillConfig = &SysConfig{
		etcdConf: &EtcdConfig{
			etcdAddr:       beego.AppConfig.String("etcdAddr"),
			timeOut:        beego.AppConfig.DefaultInt("etcdTimeOut", 10),
			etcdSecKPrefix: beego.AppConfig.String("etcdSecKeyPrefix"),
			etcdProductKey: beego.AppConfig.String("etcdProductKey"),
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
		SecretKey: beego.AppConfig.String("secretKey"),
	}
}

func initRedis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     SecKillConfig.redisConf.maxIdle,
		MaxActive:   SecKillConfig.redisConf.maxActive,
		IdleTimeout: time.Duration(SecKillConfig.redisConf.idleTimeOut) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", SecKillConfig.redisConf.redisAddr)
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
