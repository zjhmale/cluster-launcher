package etcd

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	container *EtcdContainer
}

func (suite *ContainerTestSuite) SetupSuite() {
	ctx := context.Background()
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
	suite.container = c
}

func (suite *ContainerTestSuite) TearDownSuite() {
	suite.container.Close()
}

func (suite *ContainerTestSuite) TestClientClose() {
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, &ContainerTestSuite{})
}
