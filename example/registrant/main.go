package main

import (
	"context"
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

	r := edclient.NewRegistrant(cli, "/ed-client", "/b/", &edclient.NodeInfo{
		Server: "127.0.0.1:8081",
		Tag:    make(map[string]string),
	}, 60)

	ctx, cancel := context.WithCancel(context.Background())
	err = r.Register(ctx)
	if err != nil {
		panic(err)
	}

	time.Sleep(30 * time.Second)
	r.Quit()
	cancel()
}
