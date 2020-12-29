package service

import (
	"SecLayer/conf"
	"encoding/json"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
)

type SecRes struct {
	ProductID int64
	UserID    int64
	Token     string
	Code      int
}

//redis相关操作
func initRedisPools() (err error) {
	secLayerContext.Proxy2LayerRedisPool, err = initRedisPool(conf.SecLayerSysConfig.Proxy2LayerRedis)
	secLayerContext.Layer2ProxyRedisPool, err = initRedisPool(conf.SecLayerSysConfig.Layer2ProxyRedis)
	if err != nil {
		return
	}
	return
}

//实例化redis pool
func initRedisPool(redisconf conf.RedisConf) (pool *redis.Pool, err error) {
	pool = &redis.Pool{
		MaxIdle:     redisconf.RedisMaxIdle,
		MaxActive:   redisconf.RedisMaxActive,
		IdleTimeout: time.Duration(redisconf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisconf.RedisAddr)
		},
	}

	//测试连接池
	conn := pool.Get()
	_, err = conn.Do("ping")
	if err != nil {
		return
	}
	return
}

//秒杀处理函数
func SecProcessFunc() {
	for i := 0; i < conf.SecLayerSysConfig.ReadGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go ReadHandle()
	}

	for i := 0; i < conf.SecLayerSysConfig.WriteGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go WriteHandle()
	}

	for i := 0; i < conf.SecLayerSysConfig.HandleUserGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go HandleUser()
	}
	logs.Debug("all process started")
	secLayerContext.WaitGroup.Wait()
	logs.Debug("wait all process end")
}

func ReadHandle() {
	logs.Debug("read from proxy2layer redis running")
	for {
		conn := secLayerContext.Proxy2LayerRedisPool.Get()
		for {
			data, err := redis.String(conn.Do("blpop", conf.SecLayerSysConfig.Proxy2LayerQueueName))
			if err != nil {
				logs.Error("blpop from redis failed,err:", err.Error())
				break
			}
			var reqSecReqInfo SecKillReq
			err = json.Unmarshal([]byte(data), &reqSecReqInfo)
			if err != nil {
				logs.Error("redis data json unmarshal failed,err:", err.Error())
				continue
			}

			//对于超时的请求不再处理
			if time.Now().Unix()-reqSecReqInfo.AccessTime.Unix() >= conf.SecLayerSysConfig.MaxReqWaitTime {
				logs.Warn("this req[%v] has allready expire", reqSecReqInfo)
				continue
			}

			secLayerContext.Read2HandleChan <- &reqSecReqInfo
		}
		conn.Close()
	}
}

func WriteHandle() {
	return
}

func HandleUser() {
	return
}
