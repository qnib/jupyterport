package qniblib

type Route struct {
	uid string
	target string
}

func NewRoute(uid, target string) Route {
	return Route{uid: uid, target: target}
}