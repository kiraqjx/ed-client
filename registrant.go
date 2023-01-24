package edclient

import (
	"context"
	"strconv"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

type Registrant struct {
	prefix     string
	serverName string
	nodeInfo   *NodeInfo
	leaseid    clientv3.LeaseID
	client     *clientv3.Client
	ttl        int64 // seconds
	stop       chan bool
}

func NewRegistrant(client *clientv3.Client, prefix string, serverName string,
	nodeInfo *NodeInfo, ttl int64) *Registrant {

	return &Registrant{
		prefix:     prefix,
		serverName: serverName,
		nodeInfo:   nodeInfo,
		client:     client,
		ttl:        ttl,
		stop:       make(chan bool),
	}
}

func (r *Registrant) Register(ctx context.Context) error {
	// lock key
	lockKey := stringBuilder("/lock", r.prefix, r.serverName)
	session, err := concurrency.NewSession(r.client)
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
	defer func() {
		mutex.Unlock(ctx)
	}()

	data, err := nodeInfoToByte(*r.nodeInfo)
	if err != nil {
		return err
	}
	lease, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return err
	}
	r.leaseid = lease.ID
	_, err = r.client.Put(ctx,
		stringBuilder(r.prefix, r.serverName, strconv.FormatInt(int64(r.leaseid), 10)),
		string(data),
		clientv3.WithLease(lease.ID),
	)
	if err != nil {
		return err
	}
	keepaliveChan, err := r.client.KeepAlive(ctx, r.leaseid)
	if err != nil {
		return err
	}

	go r.listener(ctx, keepaliveChan)

	return nil
}

func (r *Registrant) listener(ctx context.Context, keepaliveChan <-chan *clientv3.LeaseKeepAliveResponse) {
OUT:
	for {
		select {
		case _, ok := <-keepaliveChan:
			if !ok {
				timeChan := time.NewTicker(1 * time.Second)
				for range timeChan.C {
					r.client.Revoke(ctx, r.leaseid)
					// if revoke success, just retry register
					// if revoke error, means the lease is timeout, just retry register
					r.Register(ctx)
					break OUT
				}
			}
		case <-r.stop:
			break OUT
		}
	}
}

func (r *Registrant) Quit() error {
	_, err := r.client.Revoke(context.Background(), r.leaseid)
	if err != nil {
		return err
	}
	r.stop <- true
	return nil
}
