package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"github.com/codegangsta/cli"
	"net/http"
)

type Spawner interface {
	// Setup the spawner
	Init(ctx *cli.Context) error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user User) (map[string]Notebook, error)
	// SpawnNotebooks create a notebook
	SpawnNotebook(user User, token string, r *http.Request) (nb Notebook, err error)
	DeleteNotebook(user User, nbname string) (err error)
}
