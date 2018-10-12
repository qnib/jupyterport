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
	log.Printf("Spawner choosen: %s",ctx.String)
	switch ctx.String("backend") {
	case "kubernetes":
		spawner := qniblib.NewKubernetesSpawner()
		www.Init(&spawner, &db)
	default:
		spawner := qniblib.NewDockerSpaner()
		www.Init(&spawner, &db)
	}
	www.Start()
}



func main() {
	app := cli.NewApp()
	app.Name = "Frontend to spawn Jupyter Notebooks."
	app.Usage = "jupyterport [options]"
	app.Version = "0.1.4"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "listen-addr",
			Value: "0.0.0.0:8080",
			Usage: "IP:PORT to bind endpoint",
			EnvVar: "JUPYTERPORT_ADDR",
		},
		cli.StringFlag{
			Name:  "backend",
			Value: "docker",
			Usage: "backend to be used (docker|kubernetes)",
			EnvVar: "JUPYTERPORT_SPAWNER",
		},
		cli.StringSliceFlag{
			Name:  "docker-images",
			Value: &cli.StringSlice{"qnib/uplain-jupyter-base-notebook:2018-10-12.1", "qnib/uplain-jupyter-base-notebook:local"},
		},
		cli.BoolFlag{
			Name: "debug",
			Usage: "Be more verbose..",
			EnvVar: "JUPYTERPORT_DEBUG",
		},
	}
	app.Action = Run
	app.Run(os.Args)
}