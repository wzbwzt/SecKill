package service

import (
	"SecProxy/conf"
	"SecProxy/parameter"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/garyburd/redigo/redis"
)

var (
	SecReqChan chan *parameter.SecKillReq = make(chan *parameter.SecKillReq,
		conf.SecKillConfig.SecReqChanSize)

	UserConnMap     map[string]chan *SecResult = make(map[string]chan *SecResult, 1000)
	UserConnMapLock sync.Mutex
)

type SecResult struct {
	ProductId int
	UserId    int
	Code      int
	Token     string
}

//初始化服务
func initService() (err error) {
	err = LoadBlackList()
	if err != nil {
		return
	}

	initRedisProcessFunc()
	return
}

//将秒杀数据给写入到redis中;通过chan来实现goroutine来实现通信
func initRedisProcessFunc() {
	for i := 0; i < conf.SecKillConfig.WriteProxy2LayerGoNum; i++ {
		go WriteHanle()
	}

	for i := 0; i < conf.SecKillConfig.ReadLayer2ProxyGoNum; i++ {
		go ReadHanle()
	}

}

func WriteHanle() {
	for {
		req := <-SecReqChan
		conn := conf.Proxy2LayerRedisPool.Get()
		data, err := json.Marshal(req)
		if err != nil {
			logs.Error("req data json marshal failed,err:", err.Error())
			conn.Close()
			continue
		}
		_, err = conn.Do("LPUSH", "sec_queue", data)
		if err != nil {
			logs.Error("write data to redis failed, err:", err.Error())
			conn.Close()
			continue
		}
		conn.Close()

	}
}

func ReadHanle() {
	for {
		conn := conf.Proxy2LayerRedisPool.Get()
		replay, err := conn.Do("RPOP", "recv_queue")
		data, err := redis.String(replay, err)
		if err != nil {
			logs.Error("read from redis failed err:", err.Error())
			conn.Close()
			continue
		}
		var recData parameter.SecKillReq
		err = json.Unmarshal([]byte(data), &recData)
		if err != nil {
			logs.Error("read from redis json unmarshal failed,err:", err.Error())
			continue
		}
		userKey := fmt.Sprintf("%d_%d", recData.UserID, recData.ProductID)
		UserConnMapLock.Lock()
		_, ok := UserConnMap[userKey]
		if !ok {
			conn.Close()
			logs.Warn("user not found: %s", userKey)
			continue
		}
		SecReqChan <- &recData
		conn.Close()
	}

}

//从redis加载ip和userId的黑白名单
func LoadBlackList() (err error) {
	conn := conf.BlackRedisPool.Get()
	defer conn.Close()

	//加载人员黑名单
	replay, err := conn.Do("hgetall", "userIdBlackList")
	idlist, err := redis.Strings(replay, err)
	if err != nil {
		logs.Warn("hget failed err:", err.Error())
		return
	}
	for _, v := range idlist {
		var userID int64
		userID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			logs.Error("involid userid,id:", v)
			return
		}
		conf.SecKillConfig.UserIDBlcakMap[userID] = true
	}

	//加载IP黑名单
	replay, err = conn.Do("hgetall", "IPBlackList")
	IPlist, err := redis.Strings(replay, err)
	if err != nil {
		logs.Warn("hget failed err:", err.Error())
		return
	}
	for _, v := range IPlist {
		conf.SecKillConfig.IPBlcakMap[v] = true
	}

	//异步从reids中同步userId和Ip黑名单
	go syncBlackUserID()
	go syncBlackIP()
	return
}

func syncBlackUserID() {
	var tmp_userID []int64
	lastTime := time.Now().Unix()
	for {
		conn := conf.BlackRedisPool.Get()
		defer conn.Close()

		replay, err := conn.Do("BLPOP", "blackUserIdList", time.Second) //blpop会阻塞执行，等待时间设置为1S
		userID, err := redis.Int64(replay, err)
		if err != nil {
			logs.Warn("sync black user id[%s] failed,err:", userID, err)
			continue
		}

		tmp_userID = append(tmp_userID, userID)
		curTime := time.Now().Unix()
		if len(tmp_userID) == 100 || curTime-lastTime < 5 {
			//频繁加锁性能损耗，可以先加载到内容中，到达一定数量后再统一加载
			conf.SecKillConfig.SyncRwLock.RLock()
			for _, v := range tmp_userID {
				conf.SecKillConfig.UserIDBlcakMap[v] = true
			}
			conf.SecKillConfig.SyncRwLock.RUnlock()

			lastTime = curTime
		}
		logs.Info("sync userID list[%v] from redis success!", tmp_userID)

	}

}

func syncBlackIP() {
	var tmp_IP []string
	lastTime := time.Now().Unix()
	for {
		conn := conf.BlackRedisPool.Get()
		defer conn.Close()

		replay, err := conn.Do("BLPOP", "blackIPList", time.Second) //blpop会阻塞执行，等待时间设置为1S
		IP, err := redis.String(replay, err)
		if err != nil {
			logs.Warn("sync black ip[%s] failed,err:", IP, err)
			continue
		}

		tmp_IP = append(tmp_IP, IP)
		curTime := time.Now().Unix()
		if len(tmp_IP) == 100 || curTime-lastTime < 5 {
			//频繁加锁性能损耗，可以先加载到内容中，到达一定数量后再统一加载
			conf.SecKillConfig.SyncRwLock.RLock()
			for _, v := range tmp_IP {
				conf.SecKillConfig.IPBlcakMap[v] = true
			}
			conf.SecKillConfig.SyncRwLock.RUnlock()

			lastTime = curTime
		}
		logs.Info("sync IP list[%v] from redis success!", tmp_IP)
	}

}
