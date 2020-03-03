package etcd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/docker/go-connections/nat"
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
	listener  *EtcdListener
	endpoint  string
	dataDir   string
	context   context.Context
}

func (c *EtcdContainer) Start() {
	if c.container != nil {
		if err := c.container.Start(c.context); err != nil {
			log.Printf("Error %v when starting container: %v", err, c.endpoint)
			c.listener.FailedToStart(c, err)
		} else {
			c.listener.Started(c)
		}
	} else {
		c.listener.FailedToStart(c, errors.New("raw container not exist"))
	}
}

func (c *EtcdContainer) Stop() {
	if c.container != nil {
		if err := c.container.Terminate(c.context); err != nil {
			log.Printf("Error %v when stoping container: %v", err, c.endpoint)
		} else {
			c.listener.Stopped(c)
		}
	}
}

func (c *EtcdContainer) Restart() {
	c.Stop()
	c.Start()
}

func (c *EtcdContainer) Close() {
	c.Stop()
	c.deleteDataDir()
}

func (c *EtcdContainer) createDataDir() {
	prefix := fmt.Sprintf("etcd_cluster_mock_data_%s", c.endpoint)
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		log.Fatalf("create data directory %s failed %v", prefix, err)
	}
	c.dataDir = dir
}

func (c *EtcdContainer) deleteDataDir() {
	if err := os.RemoveAll(c.dataDir); err != nil {
		log.Fatalf("delete data directory %s failed", c.dataDir)
	}
}

func (c *EtcdContainer) getEndpoint(internalPort int) *url.URL {
	var ip string
	var port nat.Port
	var err error

	ip, err = c.container.Host(c.context)
	if err != nil {
		log.Printf("Failed to get host for container %s: %v", c.endpoint, err)
		return nil
	}
	port, err = c.container.MappedPort(c.context, (nat.Port)(fmt.Sprintf("%d", internalPort)))
	if err != nil {
		log.Printf("Failed to get %d mapped port for container %s: %v", internalPort, c.endpoint, err)
		return nil
	}

	u, err := url.Parse(fmt.Sprintf("http://%s:%s", ip, port))
	if err != nil {
		log.Printf("Failed to get url for container %v with host %s port %s", c.endpoint, ip, port)
		return nil
	}
	return u
}

func (c *EtcdContainer) ClientEndpoint() *url.URL {
	return c.getEndpoint(EtcdClientPort)
}

func (c *EtcdContainer) PeerEndpoint() *url.URL {
	return c.getEndpoint(EtcdPeerPort)
}

func NewEtcdContainer(ctx context.Context, clusterName string, listener *EtcdListener, endpoint string, endpoints []string) (*EtcdContainer, error) {
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

	ec := &EtcdContainer{
		container: c.(*tc.DockerContainer),
		listener:  listener,
		endpoint:  endpoint,
		context:   ctx,
	}
	ec.createDataDir()

	return ec, err
}
