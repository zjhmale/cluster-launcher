package etcd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"sync"
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

func NewEtcdCluster(clusterName string, nodesNum int) (*EtcdCluster, error) {
	var endpoints []string
	var containers []*EtcdContainer

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		endpoints = append(endpoints, endpoint)
	}

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)

	var rtnErr error
	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		container, err := NewEtcdContainer(ctx, wg, clusterName, listener, endpoint, endpoints)
		if err != nil {
			rtnErr = err
			break
		}
		containers = append(containers, container)
	}

	if rtnErr != nil {
		return nil, rtnErr
	}

	return &EtcdCluster{
		containers: containers,
		waitgroup:  wg,
		listener:   listener,
		context:    ctx,
	}, nil
}

func (ec *EtcdCluster) Start() error {
	for _, c := range ec.containers {
		c := c
		go func() {
			log.Printf("Starting etcd container %v", c.endpoint)
			if err := c.Start(); err != nil {
				log.Printf("Error %v when starting etcd container %v", err, c.endpoint)
			}
		}()
	}
	ec.waitgroup.Wait()
	if ec.listener.IsFailed() {
		return errors.New("Etcd cluster failed to start")
	}
	return nil
}

func (ec *EtcdCluster) Restart() error {
	for _, c := range ec.containers {
		c := c
		go func() {
			log.Printf("Restarting etcd container %v", c.endpoint)
			if err := c.Restart(); err != nil {
				log.Printf("Error %v when restarting etcd container %v", err, c.endpoint)
			}
		}()
	}
	ec.waitgroup.Wait()
	if ec.listener.IsFailed() {
		return errors.New("Etcd cluster failed to restart")
	}
	return nil
}

func (ec *EtcdCluster) Close() error {
	var rtnErr error
	for _, c := range ec.containers {
		log.Printf("Stopping etcd container %v", c.endpoint)
		if err := c.Stop(); err != nil {
			rtnErr = err
			break
		}
	}
	return rtnErr
}

func (ec *EtcdCluster) Endpoints(cb func(*EtcdContainer) (*url.URL, error)) ([]*url.URL, error) {
	var rtnErr error
	endpoints := make([]*url.URL, len(ec.containers))
	for _, c := range ec.containers {
		e, err := c.ClientEndpoint()
		if err != nil {
			rtnErr = err
			break
		} else {
			endpoints = append(endpoints, e)
		}
	}

	if rtnErr != nil {
		return nil, rtnErr
	} else {
		return endpoints, nil
	}

}

func (ec *EtcdCluster) ClientEndpoints() ([]*url.URL, error) {
	return ec.Endpoints(func(c *EtcdContainer) (*url.URL, error) { return c.ClientEndpoint() })
}

func (ec *EtcdCluster) PeerEndpoints() ([]*url.URL, error) {
	return ec.Endpoints(func(c *EtcdContainer) (*url.URL, error) { return c.PeerEndpoint() })
}
