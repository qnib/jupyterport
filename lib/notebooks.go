package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"strings"
)

type Notebook struct {
	ID 		string
	Name    string
	User    string
	Url 	string
	Token 	string
}

func NewNotebook(id, name, user, url, token string) Notebook {
	name = strings.TrimLeft(name, fmt.Sprintf("/%s_", user))
	return Notebook{ID: id, Name: name, User: user, Url: url, Token: token}
}
