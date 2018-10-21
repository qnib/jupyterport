package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"log"
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
	log.Printf("name:%s, user:%s, iurl:%s, eurl:%s, path:%s, token:%s",name, user, iurl, eurl, path, token)
	name = strings.TrimLeft(name, fmt.Sprintf("/%s_", user))
	return Notebook{
		ID: id, Spawner: spwnr,
		Name: name, User: user,
		InternalUrl: iurl, ExternalUrl: eurl,
		Path: path, Token: token,
	}
}
