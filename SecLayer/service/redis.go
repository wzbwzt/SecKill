package service

import (
	"SecLayer/conf"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
)

//redis相关操作
func initRedisPools() (err error) {
	secLayerContext = &SecLayerContext{}
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
	return
}

func WriteHandle() {
	return
}

func HandleUser() {
	return
}
