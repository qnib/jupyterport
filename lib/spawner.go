package qniblib

type Spawner interface {
	// Setup the spawner
	Init() error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user string) (map[string]Notebook, error)
	// SpawnNotebooks create a notebook
	SpawnNotebooks(user, image, token string) (err error)
}
