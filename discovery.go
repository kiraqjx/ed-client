package edclient

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

type Watcher struct {
	prefix     string
	serverName string
	Nodes      map[string]*NodeInfo
	client     *clientv3.Client
}

func NewWatcher(client *clientv3.Client, prefix string, serverName string) *Watcher {
	return &Watcher{
		prefix:     prefix,
		serverName: serverName,
		Nodes:      make(map[string]*NodeInfo),
		client:     client,
	}
}

func (w *Watcher) Start(ctx context.Context) error {
	key := stringBuilder(w.prefix, w.serverName)
	lockKey := stringBuilder("/lock", key)

	// lock key
	session, err := concurrency.NewSession(w.client)
	if err != nil {
		return err
	}
	mutex := concurrency.NewMutex(session, lockKey)

	ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	if err := mutex.Lock(ctxTimeout); err != nil {
		cancel()
		return err
	}
	cancel()

	// get old data
	gresp, err := w.client.Get(ctx, stringBuilder(w.prefix, w.serverName), clientv3.WithPrefix())
	if err != nil {
		mutex.Unlock(ctx)
		return err
	}

	kvs := gresp.Kvs
	if len(kvs) > 0 {
		for _, v := range kvs {
			w.addNode(v)
		}
	}

	// unlock key
	err = mutex.Unlock(ctx)
	if err != nil {
		return err
	}

	wch := w.client.Watch(ctx, key, clientv3.WithPrefix())
	go func() {
	OUT:
		for {
			select {
			case <-ctx.Done():
				break OUT
			case wresp := <-wch:
				for _, events := range wresp.Events {
					switch events.Type {
					case clientv3.EventTypePut:
						w.addNode(events.Kv)
					case clientv3.EventTypeDelete:
						delete(w.Nodes, string(events.Kv.Key))
					}
				}
			}
		}
	}()
	return nil
}

func (w *Watcher) addNode(kv *mvccpb.KeyValue) {
	info, err := kvToNodeInfo(kv.Value)
	if err != nil {
		fmt.Println(kv)
		log.Println(err.Error())
	}
	w.Nodes[string(kv.Key)] = info
}

func stringBuilder(args ...string) string {
	var build strings.Builder
	for _, value := range args {
		build.WriteString(value)
	}
	return build.String()
}
