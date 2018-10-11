package qniblib // import "github.com/qnib/jupyterport/lib"

type Spawner interface {
	// Setup the spawner
	Init() error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user string) (map[string]Notebook, error)
	// SpawnNotebooks create a notebook
	SpawnNotebook(user, name, port, image, token string) (nb Notebook, err error)
}
