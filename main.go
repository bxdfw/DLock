package main

import (
	"Dlock/etcd_lock"
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"49.235.108.60:2379"},
		// Endpoints: []string{"localhost:2379", "localhost:22379", "localhost:32379"}
	})

	if err != nil {
		fmt.Println("connect failed, err:", err)
	}

	defer cli.Close()

	lock := etcd_lock.NewMutex(cli, "test", 60)
	fmt.Println("lock start")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = lock.Lock(ctx)
	if err != nil {
		fmt.Println("lock failed, err:", err)
		return
	}
	fmt.Println("lock success")
}
