package etcd

import (
	"context"
	"log"

	"github.com/docker/docker/client"
	tc "github.com/testcontainers/testcontainers-go"
)

type EtcdNetwork struct {
	name    string
	context context.Context
	client  *client.Client
	network *tc.DockerNetwork
}

func NewEtcdNetwork(ctx context.Context, name string) (*EtcdNetwork, error) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	client.NegotiateAPIVersion(context.Background())

	provider, err := tc.NewDockerProvider()
	if err != nil {
		return nil, err
	}

	network, err := provider.CreateNetwork(ctx, tc.NetworkRequest{
		Name:           name,
		CheckDuplicate: true,
		SkipReaper:     true,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("Etcd network %v created", name)
	return &EtcdNetwork{
		name:    name,
		context: ctx,
		client:  client,
		network: network.(*tc.DockerNetwork),
	}, nil
}

func (en *EtcdNetwork) Remove() error {
	if err := en.client.NetworkRemove(en.context, en.network.ID); err != nil {
		return err
	}
	log.Printf("Etcd network %v removed", en.name)
	return nil
}
