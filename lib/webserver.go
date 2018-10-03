package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"github.com/kataras/go-sessions"
	"github.com/thedevsaddam/renderer"
	"log"
	"net/http"
)


var (

	cookieNameForSessionID = "mycookiesessionnameid"
	tplDir = "./tpl/*.html"
)


type Webserver struct {
	rnd 		*renderer.Render
	sess   		*sessions.Sessions
	SessionID	string
}

func NewWebserver() Webserver {
	return Webserver{}
}


func (www *Webserver) HandlerNotebooks(w http.ResponseWriter, r *http.Request) {

	// Check if user is authenticated
	sess := www.sess.Start(w, r)
	cont := NewContent(sess.GetAll())
	log.Printf("Content: %s", cont.String())
	if ! cont.Authenticated {
		http.Redirect(w, r, "/login", 303)
		return
	}
	www.rnd.HTML(w, http.StatusOK,  "notebooks", cont)

	// Print secret message
	//fmt.Fprintf(w, "%v", sess.GetAll())
}

func (www *Webserver) LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Withon LoginHandler: method:%s", r.Method)
	sess := www.sess.Start(w, r)
	cont := NewContent(sess.GetAll())
	www.rnd.HTML(w, http.StatusOK, "login-form", cont)
}
func (www *Webserver) HandlerUserLogin(w http.ResponseWriter, r *http.Request) {
	sess := www.sess.Start(w, r)
	sess.Set("authenticated", true)
	usr := r.FormValue("uname")
	sess.Set("uname", usr)
	log.Printf("User '%s' authenticated", usr)
	cont := NewContent(sess.GetAll())
	err := www.rnd.HTML(w, http.StatusOK, "user", cont)
	if err != nil {
		log.Println(err.Error())
	}
}

func (www *Webserver) HandlerHome(w http.ResponseWriter, r *http.Request) {
	sess := www.sess.Start(w, r)
	cont := NewContent(sess.GetAll())
	www.rnd.HTML(w, http.StatusOK, "home", cont)

}

func (www *Webserver) LogutHandler(w http.ResponseWriter, r *http.Request) {
	session := www.sess.Start(w, r)
	// Revoke users authentication
	session.Set("authenticated", false)
}

func (www *Webserver) Init() {
	opts := renderer.Options{
		ParseGlobPattern: tplDir,
	}

	www.rnd = renderer.New(opts)
	www.sess = sessions.New(sessions.Config{Cookie: cookieNameForSessionID})

}

func (www *Webserver) Start() {
	app := http.NewServeMux()
	app.HandleFunc("/", www.HandlerHome)
	app.HandleFunc("/notebooks", www.HandlerNotebooks)
	app.HandleFunc("/login", www.LoginFormHandler)
	app.HandleFunc("/user", www.HandlerUserLogin)
	app.HandleFunc("/logout", www.LogutHandler)

	http.ListenAndServe(":8080", app)
}
