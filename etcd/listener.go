package etcd

import (
	"sync"
)

type ContainerListener interface {
	Started(c EtcdContainer)
	FailedToStart(c EtcdContainer)
	Stopped(c EtcdContainer)
}

type EtcdListener struct{
	waitgroup sync.WaitGroup
}


