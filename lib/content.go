package qniblib

import (
	"fmt"
)

type Content struct {
	User 			string
	Authenticated 	bool
	UCPtoken		string
}

func NewContent(m map[string]interface{}) Content {
	c := Content{
		User: "Empty",
		Authenticated: false,
	}
	if v,ok := m["authenticated"]; ok {
		c.Authenticated = v.(bool)
	}
	if v, ok := m["uname"];ok {
		c.User = v.(string)
	}
	return c
}

func (c *Content) String() string {
	return fmt.Sprintf("User:%s | Auth:%v", c.User, c.Authenticated)
}
