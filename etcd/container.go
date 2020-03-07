package etcd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/docker/go-connections/nat"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
	waitgroup *sync.WaitGroup
	context   context.Context
	endpoint  string
	dataDir   string
}

func (c *EtcdContainer) Start() error {
	c.waitgroup.Add(1)
	if c.container != nil {
		if err := c.container.Start(c.context); err != nil {
			c.listener.FailedToStart(c, err)
			return err
		} else {
			c.listener.Started(c)
			return nil
		}
	} else {
		e := errors.New("raw container not exist")
		c.listener.FailedToStart(c, e)
		return e
	}
}

func (c *EtcdContainer) Stop() error {
	if c.container != nil {
		if err := c.container.Terminate(c.context); err != nil {
			return err
		} else {
			c.listener.Stopped(c)
			return nil
		}
	} else {
		return errors.New("raw container not exist")
	}
}

func (c *EtcdContainer) Restart() error {
	if err := c.Stop(); err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	return nil
}

func (c *EtcdContainer) Close() error {
	if err := c.Stop(); err != nil {
		return err
	}
	if err := c.deleteDataDir(); err != nil {
		return err
	}
	return nil
}

func (c *EtcdContainer) createDataDir() error {
	prefix := fmt.Sprintf("etcd_cluster_mock_data_%s", c.endpoint)
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return err
	}
	c.dataDir = dir
	return nil
}

func (c *EtcdContainer) deleteDataDir() error {
	if err := os.RemoveAll(c.dataDir); err != nil {
		return err
	}
	return nil
}

func (c *EtcdContainer) getEndpoint(internalPort int) (*url.URL, error) {
	var ip string
	var port nat.Port
	var err error

	ip, err = c.container.Host(c.context)
	if err != nil {
		return nil, err
	}
	port, err = c.container.MappedPort(c.context, (nat.Port)(fmt.Sprintf("%d", internalPort)))
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(fmt.Sprintf("http://%s:%s", ip, port))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (c *EtcdContainer) ClientEndpoint() (*url.URL, error) {
	return c.getEndpoint(EtcdClientPort)
}

func (c *EtcdContainer) PeerEndpoint() (*url.URL, error) {
	return c.getEndpoint(EtcdPeerPort)
}

func NewEtcdContainer(
	ctx context.Context,
	wg *sync.WaitGroup,
	clusterName string,
	listener *EtcdListener,
	endpoint string,
	endpoints []string,
) (*EtcdContainer, error) {
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

	clientTcpPort := fmt.Sprintf("%d/tcp", EtcdClientPort)
	peerTcpPort := fmt.Sprintf("%d/tcp", EtcdPeerPort)

	gcr := tc.GenericContainerRequest{
		ContainerRequest: tc.ContainerRequest{
			Image: EtcdImage,
			ExposedPorts: []string{
				clientTcpPort,
				peerTcpPort,
			},
			Cmd: cmd,
			Networks: []string{
				clusterName,
			},
			NetworkAliases: map[string][]string{
				clusterName: []string{endpoint},
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort((nat.Port)(clientTcpPort)),
				wait.ForListeningPort((nat.Port)(peerTcpPort)),
			),
			SkipReaper: true,
		},
		Started: false,
	}

	c, err := tc.GenericContainer(ctx, gcr)
	if err != nil {
		return nil, err
	}

	ec := &EtcdContainer{
		container: c.(*tc.DockerContainer),
		listener:  listener,
		waitgroup: wg,
		context:   ctx,
		endpoint:  endpoint,
	}
	if err := ec.createDataDir(); err != nil {
		return nil, err
	}

	return ec, err
}
