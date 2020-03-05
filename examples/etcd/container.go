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

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)
	c, err := NewEtcdContainer(ctx, wg, "cluster", listener, "etcd", []string{"etcd"})
	if err != nil {
		fmt.Printf("Create error %v", err)
	}
	if err := c.Start(); err != nil {
		fmt.Printf("Start error %v", err)
	}
	wg.Wait()

	go func() {
		<-stop
		if err := c.Close(); err != nil {
			fmt.Printf("Stop error %v", err)
		}
	}()

	select {}
}
