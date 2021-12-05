package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main1() {

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		logrus.Error(err)
		return
	}

	// rsp, err := client.Grant(context.TODO(), 100)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// keepalive, err := client.KeepAlive(context.TODO(), rsp.ID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// go func() {
	// 	for {
	// 		fmt.Printf("keepalive:%+v\n", <-keepalive)
	// 	}
	// }()

	go func() {
		for {
			time.Sleep(time.Second * 1)

			data := map[string]int64{
				"rand": rand.Int63n(10000000),
			}

			newStr, _ := json.Marshal(data)

			_, err := client.Put(context.TODO(), "/dwc/fish-1", string(newStr))
			if err != nil {
				fmt.Println(err)
				return
			}

		}
	}()

	go func() {

		watchCh := client.Watch(context.TODO(), "/", clientv3.WithPrefix(), clientv3.WithPrevKV())

		for {
			for res := range watchCh {
				if res.Events[0].PrevKv != nil {
					preValue := res.Events[0].PrevKv.Value
					fmt.Printf("now:%d,preValue:%+v\n", time.Now(), string(preValue))
				}

				if res.Events[0].Kv != nil {
					value := res.Events[0].Kv.Value
					fmt.Printf("now:%d,value:%+v\n", time.Now(), string(value))
				}
			}
		}
	}()

	ch := make(chan byte)
	go func() {
		time.Sleep(time.Second * 2)
		if time.Now().Unix() < 100 {
			ch <- 'a'
		}
	}()
	// client.Put(context.Background(),"/node/aa", "1234", clientv3.WithSerializable())
	<-ch
}
