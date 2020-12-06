package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

const (
	EtcdKey = "/joelWu/backend/seckill/product"
)

type SecProductInfoConf struct {
	ProductID int
	StartTime int64
	EndTime   int64
	Status    int
	Total     int
	Left      int
}

//SetLogConfToEtcd 加载数据到etcd
func SetLogConfToEtcd() {
	cli, err := etcd.New(etcd.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println("connect failed, err:", err)
		return
	}

	fmt.Println("connect succ")
	defer cli.Close()

	now := time.Now().Unix()
	var SecProductInfoConfArr []*SecProductInfoConf
	SecProductInfoConfArr = append(
		SecProductInfoConfArr,
		&SecProductInfoConf{
			ProductID: 1001,
			StartTime: now - 60,
			EndTime:   now + 3600,
			Status:    0,
			Total:     10,
			Left:      10,
		},
	)
	SecProductInfoConfArr = append(
		SecProductInfoConfArr,
		&SecProductInfoConf{
			ProductID: 1082,
			StartTime: now - 60,
			EndTime:   now + 3600,
			Status:    0,
			Total:     10,
			Left:      10,
		},
	)

	data, err := json.Marshal(SecProductInfoConfArr)
	if err != nil {
		fmt.Println("json failed, ", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = cli.Put(ctx, EtcdKey, string(data))
	if err != nil {
		fmt.Println("put failed, err:", err)
		return
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := cli.Get(ctx, EtcdKey)
	if err != nil {
		fmt.Println("get failed, err:", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
}

func main() {
	SetLogConfToEtcd()
}
