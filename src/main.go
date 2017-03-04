package main

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"context"
	"fmt"
)

func testPutAndGet() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	_, err = cli.Put(context.TODO(), "foo", "bar22")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := cli.Get(ctx, "foo")
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
	}
	// Output: foo : bar
}

func testWatchWithRange() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// watches within ['foo1', 'foo4'), in lexicographical order
	rch := cli.Watch(context.Background(), "foo1", clientv3.WithRange("foo4"))
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
	// PUT "foo1" : "bar"
	// PUT "foo2" : "bar"
	// PUT "foo3" : "bar"
}

func testWatchWithProgressNotify() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}

	rch := cli.Watch(context.Background(), "foo", clientv3.WithProgressNotify())

	wresp := <-rch
	fmt.Printf("wresp.Header.Revision: %d\n", wresp.Header.Revision)
	fmt.Println("wresp.IsProgressNotify:", wresp.IsProgressNotify())
	// wresp.Header.Revision: 0
	// wresp.IsProgressNotify: true
}

func main() {
	//testWatchWithProgressNotify()//启动了马上就返回了 所以需要sleep一段时间等待他执行结束.
	testWatchWithRange()
	//time.Sleep(10000 *time.Millisecond)
	//go testPutAndGet()
	//time.Sleep(1000 *time.Millisecond)
	//go testPutAndGet()
	//time.Sleep(1000 *time.Millisecond)
	//go testPutAndGet()
	//time.Sleep(1000000 *time.Millisecond)
	//testWatchWithRange()
}
