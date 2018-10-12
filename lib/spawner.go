package qniblib // import "github.com/qnib/jupyterport/lib"

import "net/http"

type Spawner interface {
	// Setup the spawner
	Init() error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user string) (map[string]Notebook, error)
	// SpawnNotebooks create a notebook
	SpawnNotebook(user string, r *http.Request, token string) (nb Notebook, err error)
}
