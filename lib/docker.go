package qniblib


import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"strings"
)

const (
	defaultDockerAPIVersion = "v1.37"
	baseUrl = "http://127.0.0.1"
	token = "12b755e32caa0a292f79d2615b8f973ecb2666d910d11a94"
)

var (
	ctx = context.Background()
)

type DockerSpawner struct {
	cli *client.Client
}

func NewDockerSpaner() DockerSpawner {
	return DockerSpawner{}
}

func (ds *DockerSpawner) Init() (err error){
	ds.cli, err = client.NewClientWithOpts(client.WithVersion(defaultDockerAPIVersion), client.WithHost("unix:///var/run/docker.sock"))
	if err != nil {
		panic(err)
	}
	return
}

func (ds *DockerSpawner) ListNotebooks(user string) (nbs map[string]Notebook, err error) {
	nbs = make(map[string]Notebook)
	containers, err := ds.cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		url := fmt.Sprintf("%s:%d", baseUrl, container.Ports[0].PublicPort)
		nbs[container.Names[0]] = NewNotebook(container.ID[:10], url, token)
	}
	return
}

func (ds *DockerSpawner) SpawnNotebooks(user, image, token string) (err error) {
	slice := strings.SplitAfter(image, "/")
	cntName := fmt.Sprintf("%s_%s", user, slice[len(slice)-1])
	cntCfg := container.Config{
		Env: []string{},
		Image: image,
	}
	var pm nat.PortMap
	pb := []nat.PortBinding{}
	pb = append(pb, nat.PortBinding{"0.0.0.0", ""})
	pm["8888"] = pb
	hstCfg := container.HostConfig{
		PortBindings: pm,
	}
	netCfg := network.NetworkingConfig{}
	cnt, err := ds.cli.ContainerCreate(ctx, &cntCfg, &hstCfg, &netCfg, cntName)
	_ = cnt
	return
}
