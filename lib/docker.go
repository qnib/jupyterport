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
	"net/http"
)


var (
	ctx = context.Background()
)

type DockerSpawner struct {
	Type 	string
	cli 	*client.Client
}

func NewDockerSpaner() DockerSpawner {
	return DockerSpawner{Type: "docker"}
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
		iurl := fmt.Sprintf("http://%s:%d", baseIP, container.Ports[0].PublicPort)
		eurl := fmt.Sprintf("http://%s:%d", baseIP, container.Ports[0].PublicPort)
		path := fmt.Sprintf("/user/%s/%s", user, container.Labels["name"])
		log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", container.Names[0], iurl, eurl, path)
		nbs[container.Names[0]] = NewNotebook(container.ID[:10], ds.Type, container.Names[0], user, iurl, eurl, path, token)
	}
	return
}

func (ds *DockerSpawner) SpawnNotebook(user string, r *http.Request, token string) (nb Notebook, err error) {
	cntname := r.FormValue("cntname")
	cntport := r.FormValue("cntport")
	cntimg := r.FormValue("cntimage")
	route := fmt.Sprintf("JUPYTERPORT_ROUTE=/user/%s/%s", user, cntname)
	cntName := fmt.Sprintf("%s_%s", user, cntname)
	natPrt := nat.Port(fmt.Sprintf("%d/tcp", InternalNotebookPort))
	cntCfg := container.Config{
		Env: []string{
			"JUPYTERHUB_API_TOKEN=qnib",
			route,
		},
		Image: cntimg,
		ExposedPorts: nat.PortSet{
			natPrt: {},
		},
		Labels: map[string]string{"jupyterport-user": user},
	}
	pm := make(nat.PortMap)
	pb := []nat.PortBinding{}
	pb = append(pb, nat.PortBinding{"0.0.0.0", cntport})
	pm[natPrt] = pb
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
	iurl := fmt.Sprintf("http://%s:%d", baseIP, InternalNotebookPort)
	eurl := fmt.Sprintf("http://%s:%d", baseIP, cntport)
	path := fmt.Sprintf("/user/%s/%s", user, cntname)
	log.Printf("Found notebook '%s': Internal:%s External:%s Path:%s", cntName, iurl, eurl, path)
	nb  = NewNotebook(cnt.ID[:10], ds.Type, cntName, user, iurl, eurl, path, token)
	return
}
