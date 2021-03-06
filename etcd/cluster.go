package etcd

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
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
	network    *EtcdNetwork
}

func NewEtcdCluster(clusterName string, nodesNum int) (*EtcdCluster, error) {
	var endpoints []string
	var containers []*EtcdContainer
	ctx := context.Background()

	network, err := NewEtcdNetwork(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		endpoints = append(endpoints, endpoint)
	}

	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)

	var rtnErr error
	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		container, err := NewEtcdContainer(ctx, clusterName, listener, endpoint, endpoints)
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
		network:    network,
	}, nil
}

func (ec *EtcdCluster) Trigger(action string, cb func(c *EtcdContainer) error) error {
	for _, c := range ec.containers {
		c := c
		ec.waitgroup.Add(1)
		go func() {
			log.Printf("%sing etcd container %v", strings.Title(action), c.endpoint)
			if err := cb(c); err != nil {
				log.Printf("Error %v when %sing etcd container %v", err, action, c.endpoint)
			}
		}()
	}
	ec.waitgroup.Wait()
	if ec.listener.IsFailed() {
		return fmt.Errorf("Etcd cluster failed to %s", action)
	}
	return nil
}

func (ec *EtcdCluster) Start() error {
	return ec.Trigger("start", func(c *EtcdContainer) error { return c.Start() })
}

func (ec *EtcdCluster) Restart() error {
	return ec.Trigger("restart", func(c *EtcdContainer) error { return c.Restart() })
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
	if rtnErr != nil {
		return rtnErr
	}
	if err := ec.network.Remove(); err != nil {
		return err
	}
	return nil
}

func (ec *EtcdCluster) Endpoints(cb func(*EtcdContainer) (*url.URL, error)) ([]*url.URL, error) {
	var rtnErr error
	var endpoints []*url.URL
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
