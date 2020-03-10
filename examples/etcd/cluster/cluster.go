package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	client "github.com/coreos/etcd/clientv3"
	. "github.com/zjhmale/cluster-launcher/etcd"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	clusterName := "cluster"

	cluster, err := NewEtcdCluster(clusterName, 3)
	if err != nil {
		fmt.Printf("Create cluster error %v\n", err)
	}

	if err := cluster.Start(); err != nil {
		fmt.Printf("Start cluster error %v\n", err)
	}

	endpoints, err := cluster.ClientEndpoints()
	if err != nil {
		fmt.Printf("Get client endpoints error %v\n", err)
	}

	for _, e := range endpoints {
		fmt.Printf("%v\n", e.Host)
	}
	es := []string{}
	for _, e := range endpoints {
		es = append(es, e.Host)
	}

	cli, err := client.New(client.Config{
		Endpoints:   es,
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
		if err := cluster.Close(); err != nil {
			fmt.Printf("Stop container error %v\n", err)
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
