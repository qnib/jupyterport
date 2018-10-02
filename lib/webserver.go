package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"net/http"
	"github.com/kataras/go-sessions"
)

var (
	cookieNameForSessionID = "mycookiesessionnameid"
	sess                   = sessions.New(sessions.Config{Cookie: cookieNameForSessionID})
)


type Webserver struct {}

func NewWebserver() Webserver {
	return Webserver{}
}


func (www *Webserver) SecretHandler(w http.ResponseWriter, r *http.Request) {

	// Check if user is authenticated
	ses := sess.Start(w, r)
	if auth, _ := ses.GetBoolean("authenticated"); !auth {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Print secret message
	fmt.Fprintf(w, "%v", ses.GetAll())
	w.Write([]byte("The cake is a lie!"))
}

func (www *Webserver) LoginHandler(w http.ResponseWriter, r *http.Request) {
	session := sess.Start(w, r)

	// Authentication goes here
	// ...

	// Set user as authenticated
	session.Set("authenticated", true)
}

func (www *Webserver) LogutHandler(w http.ResponseWriter, r *http.Request) {
	session := sess.Start(w, r)

	// Revoke users authentication
	session.Set("authenticated", false)
}

func (www *Webserver) Start() {
	app := http.NewServeMux()
	app.HandleFunc("/secret", www.SecretHandler)
	app.HandleFunc("/login", www.LoginHandler)
	app.HandleFunc("/logout", www.LogutHandler)

	http.ListenAndServe(":8080", app)
}
