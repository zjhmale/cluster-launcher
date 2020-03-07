package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	. "github.com/zjhmale/cluster-launcher/etcd"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	clusterName := "cluster"

	ctx := context.Background()
	network, err := NewEtcdNetwork(ctx, clusterName)
	if err != nil {
		fmt.Printf("Create network error %v", err)
	}

	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)
	c, err := NewEtcdContainer(ctx, wg, clusterName, listener, "etcd", []string{"etcd"})
	if err != nil {
		fmt.Printf("Create container error %v", err)
	}
	if err := c.Start(); err != nil {
		fmt.Printf("Start container error %v", err)
	}
	wg.Wait()

	go func() {
		<-stop
		if err := c.Close(); err != nil {
			fmt.Printf("Stop container error %v", err)
		}
		if err := network.Remove(); err != nil {
			fmt.Printf("Remove network error %v", err)
		}
		os.Exit(0)
	}()

	select {}
}
