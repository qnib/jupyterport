package qniblib

import "log"

type User struct {
	Name,UID, GID string
}

func NewUser(name,uid, gid string) User {
	log.Printf("Create User '%s' (UID:%s/GID:%s)", name, uid, gid)
	return User{name, uid, gid}
}
