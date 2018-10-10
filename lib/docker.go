package qniblib // import "github.com/qnib/jupyterport/lib"


import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"log"
)

const (
	defaultDockerAPIVersion = "v1.37"
	baseUrl = "http://127.0.0.1"
	token = "qnib"
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
	f := filters.NewArgs(filters.Arg("label", fmt.Sprintf("jupyterport-user=%s", user)))
	containers, err := ds.cli.ContainerList(ctx, types.ContainerListOptions{Filters: f})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		url := fmt.Sprintf("%s:%d", baseUrl, container.Ports[0].PublicPort)
		log.Printf("Found notebook '%s': %s", container.Names[0], url)
		nbs[container.Names[0]] = NewNotebook(container.ID[:10], container.Names[0], user, url, token)
	}
	return
}

func (ds *DockerSpawner) SpawnNotebooks(user, name, port, image, token string) (err error) {
	route := fmt.Sprintf("JUPYTERPORT_ROUTE=/user/%s/%s", user, name)
	cntName := fmt.Sprintf("%s_%s", user, name)
	cntCfg := container.Config{
		Env: []string{
			"JUPYTERHUB_API_TOKEN=qnib",
			route,
		},
		Image: image,
		ExposedPorts: nat.PortSet{
			nat.Port("8888/tcp"): {},
		},
		Labels: map[string]string{"jupyterport-user": user},
	}
	pm := make(nat.PortMap)
	pb := []nat.PortBinding{}
	pb = append(pb, nat.PortBinding{"0.0.0.0", port})
	pm["8888/tcp"] = pb
	hstCfg := container.HostConfig{PortBindings: pm}
	netCfg := network.NetworkingConfig{}
	cnt, err := ds.cli.ContainerCreate(ctx, &cntCfg, &hstCfg, &netCfg, cntName)
	if err != nil {
		log.Println(err.Error())
	}
	err = ds.cli.ContainerStart(ctx, cnt.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Println(err.Error())
	}
	return
}
