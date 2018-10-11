package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	_ "github.com/mattn/go-sqlite3"
)


type Database interface {
	// Setup the spawner
	Init() error
	// ListNotebooks returns the notebooks for a given user
	ListNotebooks(user string) (map[string]Notebook, error)
	// AddNotebook inserts a notebook with all its info
	AddNotebook(notebook Notebook) (err error)
	// RemoveNotebook removes a notebook from the DB
	RemoveNotebook(notebook Notebook) (err error)
}
