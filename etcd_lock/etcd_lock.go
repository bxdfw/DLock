package etcd_lock

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Mutex struct {
	client  *clientv3.Client
	key     string
	leaseID int64

	ttl int64
}

var defaultTTL int64 = 60

func NewMutex(client *clientv3.Client, key string, ttl int64) *Mutex {
	t := ttl
	if t <= 0 {
		t = defaultTTL
	}
	return &Mutex{client: client, key: key, ttl: t}
}

func (m *Mutex) Lock(ctx context.Context) error {
	lease := clientv3.NewLease(m.client)
	grantRsp, err := lease.Grant(ctx, m.ttl)
	if err != nil {
		return err
	}
	leaseID := grantRsp.ID
	key := fmt.Sprintf("%s/%x", m.key, leaseID)
	cmp := clientv3.Compare(clientv3.CreateRevision(m.key).WithPrefix(), "=", 0)
	put := clientv3.OpPut(key, fmt.Sprintf("%x", leaseID), clientv3.WithLease(leaseID))
	get := clientv3.OpGet(m.key, clientv3.WithLastCreate()...)
	res, err := m.client.Txn(ctx).If(cmp).Then(put).Else(get, put).Commit()
	if err != nil {
		return err
	}
	err = m.keepAlive(ctx, leaseID)
	if err != nil {
		_ = m.unLock(ctx, int64(leaseID))
		return err
	}
	if res.Succeeded {
		m.leaseID = int64(leaseID)
		return nil
	}

	lastKey := res.Responses[0].GetResponseRange().GetKvs()[0].Key
	err = m.waitDelete(ctx, string(lastKey))
	if err != nil {
		_ = m.unLock(ctx, int64(leaseID))
		return err
	}
	m.leaseID = int64(leaseID)
	return nil
}

func (m *Mutex) UnLock(ctx context.Context) error {
	return m.unLock(ctx, m.leaseID)
}
func (m *Mutex) unLock(ctx context.Context, leaseID int64) error {
	if _, err := m.client.Delete(ctx, fmt.Sprintf("%s/%x", m.key, leaseID)); err != nil {
		return err
	}
	return nil
}

func (m *Mutex) keepAlive(ctx context.Context, leaseID clientv3.LeaseID) error {
	aliveRsp, err := m.client.KeepAlive(ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}

	go func() {
		for range aliveRsp {

		}
	}()

	return nil
}

func (m *Mutex) waitDelete(ctx context.Context, key string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	watchChan := m.client.Watch(ctx, key)
	var watchResp clientv3.WatchResponse
	for watchResp = range watchChan {
		for _, event := range watchResp.Events {
			if event.Type == mvccpb.DELETE {
				return nil
			}
		}
	}

	return fmt.Errorf("watch for wait delete %s error", m.key)
}

func (m *Mutex) Key() string {
	return m.key
}

func (m *Mutex) TTL() int64 {
	return m.ttl
}

func (m *Mutex) SetTTL(ttl int64) {
	if ttl <= 0 {
		m.ttl = defaultTTL
	} else {
		m.ttl = ttl
	}
}
