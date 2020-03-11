package etcd

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	container *EtcdContainer
	network   *EtcdNetwork
	client    *client.Client
}

func (suite *ContainerTestSuite) SetupSuite() {
	clusterName := "cluster"
	ctx := context.Background()
	network, err := NewEtcdNetwork(ctx, clusterName)
	if err != nil {
		suite.T().Fatalf("Create network error %v\n", err)
	}

	wg := &sync.WaitGroup{}
	listener := NewEtcdListener(wg)
	c, err := NewEtcdContainer(ctx, "cluster", listener, "etcd", []string{"etcd"})
	if err != nil {
		suite.T().Fatalf("Error %v when creating etcd container", err)
	}

	wg.Add(1)
	if err := c.Start(); err != nil {
		suite.T().Fatalf("Error %v when starting etcd container", err)
	}
	wg.Wait()

	endpoint, err := c.ClientEndpoint()
	if err != nil {
		fmt.Printf("Get client endpoint error %v\n", err)
	}

	cli, err := client.New(client.Config{
		Endpoints:   []string{endpoint.Host},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		fmt.Printf("Create etcd client error %v\n", err)
	}

	suite.container = c
	suite.network = network
	suite.client = cli
}

func (suite *ContainerTestSuite) TearDownSuite() {
	if err := suite.container.Close(); err != nil {
		suite.T().Fatalf("Error %v when closing etcd container", err)
	}
	if err := suite.network.Remove(); err != nil {
		suite.T().Fatalf("Error %v when removing docker network", err)
	}
	if err := suite.client.Close(); err != nil {
		suite.T().Fatalf("Error %v when closing client", err)
	}
}

func (suite *ContainerTestSuite) TestClientPutGet() {
	key := "key"
	val := "xyz"
	kv := client.NewKV(suite.client)
	_, err := kv.Put(context.TODO(), key, val)
	if err != nil {
		suite.T().Fatalf("Put kv error %v\n", err)
	}

	resp, err := kv.Get(context.TODO(), "key")
	if err != nil {
		suite.T().Fatalf("Get kv error %v\n", err)
	}
	pair := resp.Kvs[0]
	if string(pair.Key) != key || string(pair.Value) != val {
		suite.T().Fatalf("Data inconsistent: Key %s Value %s\n", pair.Key, pair.Value)
	}
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, &ContainerTestSuite{})
}
