package edclient

import (
	"context"
	"encoding/json"
	"sync"
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

	value := <-watcher.ChangeEvent()
	assert.Equal(t, true, value)
	if len(watcher.Nodes) > 0 {
		watchData := watcher.Nodes[key]
		assert.Equal(t, *watchData, *testNodeInfo)
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

	<-watcher.ChangeEvent()

	if len(watcher.Nodes) > 0 {
		for _, value := range watcher.Nodes {
			assert.Equal(t, *value, *nodeInfo)
		}
	} else {
		t.Error()
	}

	registrant.Quit()

	<-watcher.ChangeEvent()
	if len(watcher.Nodes) != 0 {
		t.Error()
	}
}

func Test_Lb(t *testing.T) {
	nodes := []*NodeInfo{
		{
			Server: "127.0.0.1:8081",
		},
		{
			Server: "127.0.0.1:8082",
		},
		{
			Server: "127.0.0.1:8083",
		},
	}

	lb := NewLb(nodes)
	lbMap := make(map[string]*NodeInfo)

	var wg sync.WaitGroup
	wg.Add(3)

	var lock sync.Mutex

	go func() {
		node1 := lb.Lb()
		lock.Lock()
		lbMap[node1.Server] = node1
		lock.Unlock()
		wg.Done()
	}()

	go func() {
		node2 := lb.Lb()
		lock.Lock()
		lbMap[node2.Server] = node2
		lock.Unlock()
		wg.Done()
	}()

	go func() {
		node3 := lb.Lb()
		lock.Lock()
		lbMap[node3.Server] = node3
		lock.Unlock()
		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, len(lbMap), 3)
}

func Test_Lb_Map(t *testing.T) {
	nodes := make(map[string]*NodeInfo)
	nodes["1"] = &NodeInfo{
		Server: "127.0.0.1:8081",
	}
	nodes["2"] = &NodeInfo{
		Server: "127.0.0.1:8082",
	}
	nodes["3"] = &NodeInfo{
		Server: "127.0.0.1:8083",
	}

	lb := NewLbFromMap(nodes)
	lbMap := make(map[string]*NodeInfo)

	var wg sync.WaitGroup
	wg.Add(3)

	var lock sync.Mutex

	go func() {
		node1 := lb.Lb()
		lock.Lock()
		lbMap[node1.Server] = node1
		lock.Unlock()
		wg.Done()
	}()

	go func() {
		node2 := lb.Lb()
		lock.Lock()
		lbMap[node2.Server] = node2
		lock.Unlock()
		wg.Done()
	}()

	go func() {
		node3 := lb.Lb()
		lock.Lock()
		lbMap[node3.Server] = node3
		lock.Unlock()
		wg.Done()
	}()

	wg.Wait()

	assert.Equal(t, len(lbMap), 3)
}
