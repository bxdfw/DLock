# DLock
Non-recursive Distributed Lock Implementation Based on etcd

Usage
`func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
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
}`
