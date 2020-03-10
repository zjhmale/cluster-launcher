package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	client "github.com/coreos/etcd/clientv3"
	. "github.com/zjhmale/cluster-launcher/etcd"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	clusterName := "cluster"

	ctx := context.Background()
	network, err := NewEtcdNetwork(ctx, clusterName)
	if err != nil {
		fmt.Printf("Create network error %v\n", err)
	}

	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)
	wg.Add(1)
	c, err := NewEtcdContainer(ctx, clusterName, listener, "etcd", []string{"etcd"})
	if err != nil {
		fmt.Printf("Create container error %v\n", err)
	}
	if err := c.Start(); err != nil {
		fmt.Printf("Start container error %v\n", err)
	}
	wg.Wait()

	endpoint, err := c.ClientEndpoint()
	if err != nil {
		fmt.Printf("Get client endpoint error %v\n", err)
	}

	cli, err := client.New(client.Config{
		Endpoints:   []string{endpoint.Host},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Printf("Create etcd client error %v\n", err)
	}

	kv := client.NewKV(cli)
	_, err = kv.Put(context.TODO(), "key", "xyz")
	if err != nil {
		fmt.Printf("Put kv error %v\n", err)
	}

	resp, err := kv.Get(context.TODO(), "key")
	if err != nil {
		fmt.Printf("Get kv error %v\n", err)
	}
	pair := resp.Kvs[0]
	fmt.Printf("Key %s Value %s\n", pair.Key, pair.Value)

	go func() {
		<-stop
		if err := c.Close(); err != nil {
			fmt.Printf("Stop container error %v\n", err)
			os.Exit(1)
		}
		if err := network.Remove(); err != nil {
			fmt.Printf("Remove network error %v\n", err)
			os.Exit(1)
		}
		if err := cli.Close(); err != nil {
			fmt.Printf("Stop client error %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	select {}
}
