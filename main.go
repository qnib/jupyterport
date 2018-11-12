package main

import (
	"github.com/codegangsta/cli"
	"github.com/qnib/jupyterport/lib"
	"log"
	"os"
)


func Run(ctx *cli.Context) {
	www := qniblib.NewWebserver(ctx)
	db := qniblib.NewLocalDB()
	log.Printf("Spawner choosen: %s",ctx.String("backend"))
	switch ctx.String("backend") {
	case "kubernetes":
		spawner := qniblib.NewKubernetesSpawner()
		www.Init(&spawner, &db)
	default:
		log.Panicf("Unkown spawner: %s")
	}
	www.Start()
}



func main() {
	app := cli.NewApp()
	app.Name = "Frontend to spawn Jupyter Notebooks."
	app.Usage = "jupyterport [options]"
	app.Version = "0.2.4"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen-addr",
			Value: "0.0.0.0:8080",
			Usage: "IP:PORT to bind endpoint",
			EnvVar: "JUPYTERPORT_ADDR",
		}, cli.StringFlag{
			Name:  "backend",
			Value: "docker",
			Usage: "backend to be used (docker|kubernetes)",
			EnvVar: "JUPYTERPORT_SPAWNER",
		}, cli.StringFlag{
			Name:  "k8s-namespace",
			Value: "default",
			Usage: "Kubernetes namespace",
			EnvVar: "JUPYTERPORT_KUBERNETES_NAMESPACE",
		}, cli.StringSliceFlag{
			Name:  "jupyter-images",
			Value: &cli.StringSlice{"qnib/uplain-jupyter-base-notebook:latest"},
			EnvVar: "JUPYTERPORT_JUPYTER_IMAGES",
		}, cli.StringSliceFlag{
			Name:  "notebook-images",
			Value: &cli.StringSlice{"qnib/jupyter-notebooks:latest"},
			EnvVar: "JUPYTERPORT_NOTEBOOK_IMAGES",
		}, cli.StringSliceFlag{
			Name:  "data-images",
			Value: &cli.StringSlice{"qnib/data-cifar10:latest"},
			EnvVar: "JUPYTERPORT_DATA_IMAGES",
		}, cli.StringFlag{
			Name:  "external-address",
			Value: "127.0.0.1",
			Usage: "External address of services",
			EnvVar: "JUPYTERPORT_EXT_ADDR",
		}, cli.BoolFlag{
			Name: "debug",
			Usage: "Be more verbose..",
			EnvVar: "JUPYTERPORT_DEBUG",
		},
	}
	app.Action = Run
	app.Run(os.Args)
}