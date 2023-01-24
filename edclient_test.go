package edclient

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/assert"
)

func Test_Discovery(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		t.Error(err)
	}

	// clear data
	cli.Delete(context.Background(), "/test-discovery", clientv3.WithPrefix())
	defer func() {
		cli.Delete(context.Background(), "/test-discovery", clientv3.WithPrefix())
	}()

	// start watcher
	watcher := NewWatcher(cli, "/test-discovery", "/test")
	err = watcher.Start(context.Background())
	if err != nil {
		t.Error(err)
	}

	// add test data
	testNodeInfo := &NodeInfo{
		Server: "127.0.0.1:8081",
		Tag:    make(map[string]string),
	}
	valueJson, err := json.Marshal(testNodeInfo)
	if err != nil {
		t.Error(err)
	}
	resp, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		t.Error(err)
	}
	key := "/test-discovery/test/1"
	cli.Put(context.Background(), key, string(valueJson), clientv3.WithLease(resp.ID))

OUT:
	for {
		if len(watcher.Nodes) > 0 {
			watchData := watcher.Nodes[key]
			assert.Equal(t, *watchData, *testNodeInfo)
			break OUT
		}
	}
}

func Test_Registrant(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		t.Error(err)
	}

	// start watcher
	watcher := NewWatcher(cli, "/test-registrant", "/test")
	err = watcher.Start(context.Background())
	if err != nil {
		t.Error(err)
	}

	nodeInfo := &NodeInfo{
		Server: "127.0.0.1:8081",
		Tag:    make(map[string]string),
	}
	registrant := NewRegistrant(cli, "/test-registrant", "/test", nodeInfo, 30)

	err = registrant.Register(context.Background())
	if err != nil {
		t.Error(err)
	}

OUT:
	for {
		if len(watcher.Nodes) > 0 {
			for _, value := range watcher.Nodes {
				assert.Equal(t, *value, *nodeInfo)
			}
			break OUT
		}
	}

	registrant.Quit()

	for {
		if len(watcher.Nodes) == 0 {
			break
		}
	}
}
