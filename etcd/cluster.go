package etcd

import (
	"fmt"
	"net/url"
)

type ContainerCluster interface {
	Start()
	Restart()
	Close()
	clientEndpoints() []*url.URL
	peerEndpoints() []*url.URL
}

type EtcdCluster struct {
	containers []*EtcdContainer
}

func NewEtcdCluster(clusterName string, nodesNum int) *EtcdCluster {
	var endpoints []string
	var containers []*EtcdContainer

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		endpoints = append(endpoints, endpoint)
	}

	for i := 0; i < nodesNum; i++ {
		endpoint := fmt.Sprintf("etcd%d", i)
		container, err := NewEtcdContainer(clusterName, endpoint, endpoints)
		if err != nil {
			continue
		}
		containers = append(containers, container)
	}

	return &EtcdCluster{containers: containers}
}

func (ec *EtcdCluster) Start() {

}

func (ec *EtcdCluster) Restart() {

}

func (ec *EtcdCluster) Close() {

}

