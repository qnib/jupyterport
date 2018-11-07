package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"strings"
)

type Content struct {
	User 			User
	Authenticated 	bool
	UCPtoken		string
	Notebooks		map[string]Notebook
	JupyterImages 	[]DockerImage
	NotebookImages 	[]DockerImage
	DataImages 	[]DockerImage
}

func NewContent(m map[string]interface{}) Content {
	c := Content{
		User: User{Name: "Empty"},
		Authenticated: false,
	}
	if v,ok := m["authenticated"]; ok {
		c.Authenticated = v.(bool)
	}
	if v, ok := m["uname"];ok {
		c.User = User{Name: v.(string)}
	}
	return c
}

func (c *Content) String() string {
	jimgs := []string{}
	for _, img := range c.JupyterImages {
		jimgs = append(jimgs, img.String())
	}
	nbimgs := []string{}
	for _, img := range c.NotebookImages {
		nbimgs = append(nbimgs, img.String())
	}
	return fmt.Sprintf("User:%s | Auth:%v | JupyterImages:%s | NotebookImages:%s", c.User, c.Authenticated, strings.Join(jimgs, ","), strings.Join(nbimgs, ","))
}
