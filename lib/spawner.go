package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"github.com/codegangsta/cli"
	"net/http"
)

type Spawner interface {
	// Setup the spawner
	Init(ctx *cli.Context) error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user, extAddr string) (map[string]Notebook, error)
	// SpawnNotebooks create a notebook
	SpawnNotebook(user string, r *http.Request, token, extAddr string) (nb Notebook, err error)
}
