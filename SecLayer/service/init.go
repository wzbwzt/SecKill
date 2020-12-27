package service

import (
	"SecLayer/conf"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	etcd "go.etcd.io/etcd/clientv3"
)

var (
	secLayerContext    *SecLayerContext
	MapSecKillProducts = make(map[int]*conf.SecProductInfoConf)
)

type SecLayerContext struct {
	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	EtcdClient           *etcd.Client
	WaitGroup            sync.WaitGroup
}

func InitSecLayer() (err error) {
	//初始化redis,生成实例
	err = initRedisPools()
	if err != nil {
		logs.Error("init redis pool failed,err:", err.Error())
		return
	}
	//初始化etcd
	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed ,err:", err.Error())
		return
	}

	//从etcd中加载商品信息
	err = loadProductFromEtcd()
	if err != nil {
		logs.Error(" load product from etcd failed,err:", err.Error())
		return
	}

	//检测更新商品信息
	initSecProductWatcher()
	return
}

func initEtcd() (err error) {
	secLayerContext.EtcdClient, err = etcd.New(etcd.Config{
		Endpoints:   []string{conf.SecLayerSysConfig.EtcdConfig.EtcdAddr},
		DialTimeout: time.Duration(conf.SecLayerSysConfig.EtcdConfig.Timeout) * time.Millisecond,
	})

	if err != nil {
		return
	}
	return
}

//从etcd中加载商品配置信息
func loadProductFromEtcd() (err error) {
	key := fmt.Sprintf("%s/%s", conf.SecLayerSysConfig.EtcdConfig.EtcdSecKeyPrefix,
		conf.SecLayerSysConfig.EtcdConfig.EtcdSecProductKey)
	rsp, err := secLayerContext.EtcdClient.Get(context.Background(), key)
	if err != nil {
		return
	}

	var secProductInfo []conf.SecProductInfoConf
	for k, v := range rsp.Kvs {
		logs.Debug("key:%s value:%s", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			return
		}
		logs.Debug("secInfo conf is %v", secProductInfo)
	}
	updateSecProductInfo(secProductInfo)

	return
}

//更新商品信息到内存中
func updateSecProductInfo(confs []conf.SecProductInfoConf) {
	//每个加锁效率会是问题，可以先放入到一个临时变量中，再加锁赋值
	tmp := make(map[int]*conf.SecProductInfoConf)
	for _, v := range confs {
		ttmp := v
		tmp[v.ProductId] = &ttmp
	}

	conf.SecLayerSysConfig.SecProductInfoUpdateLock.RLock()
	MapSecKillProducts = tmp
	conf.SecLayerSysConfig.SecProductInfoUpdateLock.RUnlock()
}

//监听秒杀配置是否改变
func initSecProductWatcher() {
	key := fmt.Sprintf("%s/%s", conf.SecLayerSysConfig.EtcdConfig.EtcdSecKeyPrefix,
		conf.SecLayerSysConfig.EtcdConfig.EtcdSecProductKey)
	go watchSecProductKey(key)
}

//检测更新etcd中的商品信息
func watchSecProductKey(key string) {
	logs.Debug("begin watch key: %s", key)

	for {
		watchChan := secLayerContext.EtcdClient.Watch(context.Background(), key)
		var secProductInfo []conf.SecProductInfoConf
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
