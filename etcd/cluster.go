package etcd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/tevino/abool"
)

type ContainerCluster interface {
	Start()
	Restart()
	Close()
	ClientEndpoints() []*url.URL
	PeerEndpoints() []*url.URL
}

type EtcdCluster struct {
	containers []*EtcdContainer
	waitgroup  *sync.WaitGroup
	listener   *EtcdListener
	context    context.Context
}

func NewEtcdCluster(clusterName string, nodesNum int) *EtcdCluster {
	var endpoints []string
	var containers []*EtcdContainer

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		endpoints = append(endpoints, endpoint)
	}

	ctx := context.Background()
	wg := sync.WaitGroup{}
	listener := &EtcdListener{
		waitgroup:     &wg,
		failedToStart: abool.New(),
	}

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		container, err := NewEtcdContainer(ctx, clusterName, listener, endpoint, endpoints)
		if err != nil {
			continue
		}
		containers = append(containers, container)
	}

	return &EtcdCluster{
		containers: containers,
		waitgroup:  &wg,
		listener:   listener,
		context:    ctx,
	}
}

func (ec *EtcdCluster) Start() {
	for _, c := range ec.containers {
		c := c
		ec.waitgroup.Add(1)
		go func() {
			log.Printf("Starting etcd container %v", c.endpoint)
			c.Start()
		}()
	}
	ec.waitgroup.Wait()
	if ec.listener.failedToStart.IsSet() {
		log.Fatal("Etcd cluster failed to start")
	}
}

func (ec *EtcdCluster) Restart() {
	for _, c := range ec.containers {
		c := c
		ec.waitgroup.Add(1)
		go func() {
			log.Printf("Restarting etcd container %v", c.endpoint)
			c.Restart()
		}()
	}
	ec.waitgroup.Wait()
	if ec.listener.failedToStart.IsSet() {
		log.Fatal("Etcd cluster failed to restart")
	}
}

func (ec *EtcdCluster) Close() {
	for _, c := range ec.containers {
		log.Printf("Stopping etcd container %v", c.endpoint)
		c.Stop()
	}
}

func (ec *EtcdCluster) ClientEndpoints() []*url.URL {
	endpoints := make([]*url.URL, len(ec.containers))
	for _, c := range ec.containers {
		endpoints = append(endpoints, c.ClientEndpoint())
	}
	return endpoints
}

func (ec *EtcdCluster) PeerEndpoints() []*url.URL {
	endpoints := make([]*url.URL, len(ec.containers))
	for _, c := range ec.containers {
		endpoints = append(endpoints, c.PeerEndpoint())
	}
	return endpoints
}
