package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"strings"
)

type Notebook struct {
	ID 			string
	Spawner		string
	Name    	string
	User    	string
	InternalUrl string
	ExternalUrl string
	Path		string
	Token 		string
}

func NewNotebook(id, spwnr, name, user, iurl, eurl, path, token string) Notebook {
	name = strings.TrimLeft(name, fmt.Sprintf("/%s_", user))
	return Notebook{
		ID: id, Spawner: spwnr,
		Name: name, User: user,
		InternalUrl: iurl, ExternalUrl: eurl,
		Path: path, Token: token,
	}
}
