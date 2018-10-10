package qniblib // import "github.com/qnib/jupyterport/lib"

type Route struct {
	uid string
	target string
}

func NewRoute(uid, target string) Route {
	return Route{uid: uid, target: target}
}