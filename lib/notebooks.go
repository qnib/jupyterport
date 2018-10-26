package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"log"
	"strings"
)

type Notebook struct {
	ID 			string
	Status		string
	Spawner		string
	Name    	string
	User    	string
	InternalUrl string
	Path		string
	Token 		string
}

func NewNotebook(id, spwnr, name, user, iurl, path, token, status string) Notebook {
	log.Printf("name:%s, user:%s, iurl:%s, path:%s, token:%s / status:%s",name, user, iurl, path, token, status)
	name = strings.TrimLeft(name, fmt.Sprintf("/%s_", user))
	return Notebook{
		ID: id, Spawner: spwnr,
		Name: name, User: user,
		InternalUrl: iurl,
		Path: path, Token: token,
		Status: status,
	}
}
