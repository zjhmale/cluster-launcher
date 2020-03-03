package etcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	tc "github.com/testcontainers/testcontainers-go"
)

const (
	EtcdImage      = "gcr.io/etcd-development/etcd:v3.3"
	EtcdClientPort = 2379
	EtcdPeerPort   = 2380
	EtcdDataDir    = "/data.etcd"
)

type EtcdContainer struct {
	container *tc.DockerContainer
	name      string
	dataDir   string
}

func (c *EtcdContainer) Start(ctx context.Context) {
	if c.container != nil {
		if err := c.container.Start(ctx); err != nil {
			if name, e := c.container.Name(ctx); e != nil {
				log.Printf("Error when starting container: %v", name)
			}
		}

	}
}

func (c *EtcdContainer) Stop(ctx context.Context) {
	if c.container != nil {
		if err := c.container.Terminate(ctx); err != nil {
			if name, e := c.container.Name(ctx); e != nil {
				log.Printf("Error when stoping container: %v", name)
			}
		}
	}
}

func (c *EtcdContainer) Restart(ctx context.Context) {
	c.Stop(ctx)
	c.Start(ctx)
}

func (c *EtcdContainer) Close(ctx context.Context) {
	c.Stop(ctx)
	c.deleteDataDir()
}

func (c *EtcdContainer) createDataDir() {
	prefix := fmt.Sprintf("etcd_cluster_mock_data_%s", c.name)
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Fatalf("create data directory %s failed", prefix)
	}
	c.dataDir = dir
}

func (c *EtcdContainer) deleteDataDir() {
	if err := os.RemoveAll(c.dataDir); err != nil {
		log.Fatalf("delete data directory %s failed", c.dataDir)
	}
}

func NewEtcdContainer(clusterName string, endpoint string, endpoints []string) (*EtcdContainer, error) {
	clientUrl := fmt.Sprintf("http://0.0.0.0:%d", EtcdClientPort)
	cmd := []string{
		"etcd",
		"--name", endpoint,
		"--advertise-client-urls", clientUrl,
		"--listen-client-urls", clientUrl,
		"--data-dir", EtcdDataDir,
	}

	if len(endpoints) > 0 {
		clusterEndpoints := []string{}
		for _, e := range endpoints {
			clusterEndpoints = append(clusterEndpoints, fmt.Sprintf("%s=http://%s:%d", e, e, EtcdPeerPort))
		}

		cmd = append(
			cmd,
			"--initial-advertise-peer-urls", fmt.Sprintf("http://%s:%d", endpoint, EtcdPeerPort),
			"--listen-peer-urls", fmt.Sprintf("http://0.0.0.0:%d", EtcdPeerPort),
			"--initial-cluster-token", clusterName,
			"--initial-cluster", strings.Join(clusterEndpoints, ","),
			"--initial-cluster-state", "new",
		)
	}

	ctx := context.Background()
	req := tc.ContainerRequest{
		Image: EtcdImage,
		ExposedPorts: []string{
			fmt.Sprintf("%d/tcp", EtcdClientPort),
			fmt.Sprintf("%d/tcp", EtcdPeerPort),
		},
		Cmd: cmd,
		Networks: []string{
			clusterName,
		},
		NetworkAliases: map[string][]string{
			clusterName: []string{endpoint},
		},
	}

	c, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	ec := &EtcdContainer{container: c.(*tc.DockerContainer), name: endpoint}
	ec.createDataDir()

	return ec, err
}
