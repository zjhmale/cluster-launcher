package etcd

import (
	"log"
	"sync"

	"github.com/tevino/abool"
)

type ContainerListener interface {
	Started(c *EtcdContainer)
	FailedToStart(c *EtcdContainer, e error)
	Stopped(c *EtcdContainer)
}

type EtcdListener struct {
	waitgroup     sync.WaitGroup
	failedToStart *abool.AtomicBool
}

func (el *EtcdListener) Started(c *EtcdContainer) {
	log.Printf("Etcd container %v started", c.endpoint)
	el.waitgroup.Done()
}

func (el *EtcdListener) FailedToStart(c *EtcdContainer, e error) {
	log.Printf("Etcd container %v start failed %v", c.endpoint, e.Error())
	el.failedToStart.Set()
	el.waitgroup.Done()
}

func (el *EtcdListener) Stopped(c *EtcdContainer) {
	log.Printf("Etcd container %v stopped", c.endpoint)
}
