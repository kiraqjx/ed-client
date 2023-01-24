package main

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	edclient "github.com/kiraqjx/ed-client"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	watcher := edclient.NewWatcher(cli, "/ed-client", "/b/")

	ctx, cancel := context.WithCancel(context.Background())
	err = watcher.Start(ctx)
	if err != nil {
		panic(err)
	}

	time.Sleep(15 * time.Second)
	fmt.Println("watch node:", watcher.Nodes)
	time.Sleep(15 * time.Second)
	fmt.Println("watch node:", watcher.Nodes)
	cancel()
}
