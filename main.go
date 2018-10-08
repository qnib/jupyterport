package main

import (
	"github.com/qnib/jupyterport/lib"
)


func main() {
	spawner := qniblib.NewDockerSpaner()
	www := qniblib.NewWebserver()
	www.Init(&spawner)
	www.Start()
}

