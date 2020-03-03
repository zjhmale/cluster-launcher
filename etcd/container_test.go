package etcd

import (
	"sync"
	"context"
	"testing"
	"github.com/stretchr/testify/suite"
	"github.com/tevino/abool"
)

type ContainerTestSuite struct {
	suite.Suite
	container *EtcdContainer
}

func (suite *ContainerTestSuite) SetupSuite() {
	ctx := context.Background()
	listener := &EtcdListener{
		waitgroup: &sync.WaitGroup{},
		failedToStart: abool.New(),
	}
	c, err := NewEtcdContainer(ctx, "cluster", listener, "etcd", []string{"etcd"})
	if err != nil {
		suite.T().Fatalf("Error %v when starting etcd container", err)
	}
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
