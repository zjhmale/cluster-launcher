package etcdclusterlauncher

import (
)

type ContainerListener interface {
	Started(c EtcdContainer)
	FailedToStart(c EtcdContainer)
	Stopped(c EtcdContainer)
}
