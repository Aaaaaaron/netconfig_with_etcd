package main

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"fmt"
	"context"
	"time"
)

var (
	endpoints      = []string{"localhost:2379"}
	dialTimeout    = 50 * time.Second
	requestTimeout = 10000 * time.Millisecond
)

func Put(configKey, configValue string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.TODO(), configKey, configValue)
	if err != nil {
		log.Fatal(err)
	}
}

func WatchWithRange(startKey, endKey string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// watches within ['startKey', 'endKey'), in lexicographical order
	rch := cli.Watch(context.Background(), startKey, clientv3.WithRange(endKey))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

func WatchWithPrefix() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	rch := cli.Watch(context.Background(), "waf", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

func main() {
	WatchWithPrefix()
	//time.Sleep(10000*time.Millisecond)
	//go Put("waf_ip link set down ", "eth1")
}
