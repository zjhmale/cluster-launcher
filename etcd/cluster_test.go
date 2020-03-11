package etcd

import (
	"context"
	"testing"
	"time"

	client "github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/suite"
)

type ClusterTestSuite struct {
	suite.Suite
	cluster *EtcdCluster
	client  *client.Client
}

func (suite *ClusterTestSuite) SetupSuite() {
	clusterName := "cluster"
	cluster, err := NewEtcdCluster(clusterName, 3)
	if err != nil {
		suite.T().Fatalf("Create cluster error %v\n", err)
	}

	if err := cluster.Start(); err != nil {
		suite.T().Fatalf("Start cluster error %v\n", err)
	}

	endpoints, err := cluster.ClientEndpoints()
	if err != nil {
		suite.T().Fatalf("Get client endpoints error %v\n", err)
	}

	es := []string{}
	for _, e := range endpoints {
		es = append(es, e.Host)
	}

	cli, err := client.New(client.Config{
		Endpoints:   es,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		suite.T().Fatalf("Create etcd client error %v\n", err)
	}

	suite.cluster = cluster
	suite.client = cli
}

func (suite *ClusterTestSuite) TearDownSuite() {
	if err := suite.cluster.Close(); err != nil {
		suite.T().Fatalf("Error %v when closing etcd container", err)
	}
	if err := suite.client.Close(); err != nil {
		suite.T().Fatalf("Error %v when closing client", err)
	}
}

func (suite *ClusterTestSuite) TestClientPutGet() {
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

func TestClusterTestSuite(t *testing.T) {
	suite.Run(t, &ClusterTestSuite{})
}
