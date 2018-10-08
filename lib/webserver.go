package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kataras/go-sessions"
	"github.com/thedevsaddam/renderer"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)


var (

	cookieNameForSessionID = "mycookiesessionnameid"
	tplDir = "./tpl/*.html"
)


type Webserver struct {
	rnd 		*renderer.Render
	sess   		*sessions.Sessions
	revProx		map[string]http.Handler
	SessionID	string
	router		*mux.Router
	routeChan   chan Route
	spawner 	Spawner
}

func NewWebserver() Webserver {
	return Webserver{
		routeChan: make(chan Route),
	}
}



func (www *Webserver) HandlerNotebooks(w http.ResponseWriter, r *http.Request) {
	var err error
	// Check if user is authenticated
	sess := www.sess.Start(w, r)
	cont := NewContent(sess.GetAll())
	cont.Notebooks, err = www.ListNotebooks(cont.User)
	if err != nil {
		log.Println(err.Error())
		cont.Notebooks = make(map[string]Notebook)
		www.rnd.HTML(w, http.StatusOK,  "notebooks", cont)
		return
	}
	//cont.Notebooks["note1"] = Notebook{ID: "test", Url: "/user/test", Token: "12b755e32caa0a292f79d2615b8f973ecb2666d910d11a94"}
	log.Printf("Content: %s", cont.String())
	if ! cont.Authenticated {
		http.Redirect(w, r, "/login", 303)
		return
	}
	www.rnd.HTML(w, http.StatusOK,  "notebooks", cont)
}

func (www *Webserver) ListNotebooks(user string) (nbs map[string]Notebook, err error) {
	return www.spawner.ListNotebooks(user)
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
	log.Printf("Add route for user %s", usr)
	err = www.AddRoute(usr, "http://127.0.0.1:8888/tree?token=12b755e32caa0a292f79d2615b8f973ecb2666d910d11a94")
	if err != nil {
		log.Println(err.Error())
	}
}

func (www *Webserver) HandlerStartContainer(w http.ResponseWriter, r *http.Request) {
	sess := www.sess.Start(w, r)
	www.spawner.SpawnNotebooks(sess.GetString("uname"), r.FormValue("cntimage"), token)
	cont := NewContent(sess.GetAll())
	www.rnd.HTML(w, http.StatusOK, "home", cont)
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

func (www *Webserver) Init(spawner Spawner) {
	opts := renderer.Options{
		ParseGlobPattern: tplDir,
	}
	spawner.Init()
	www.spawner = spawner
	www.router = mux.NewRouter()
	www.rnd = renderer.New(opts)
	www.sess = sessions.New(sessions.Config{Cookie: cookieNameForSessionID})

}

func (www *Webserver) AddRoute(uid, target string) (err error) {
	remote, err := url.Parse(target)
	if err != nil {
		return
	}
	prxy := httputil.NewSingleHostReverseProxy(remote)
	link := fmt.Sprintf("/user/%s", uid)
	log.Printf("%s -> %s", link, target)
	www.router.HandleFunc(link, handler(prxy))
	return
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}

func (www *Webserver) Start() {
	// Forward user notebooks
	www.router.HandleFunc("/", www.HandlerHome)
	www.router.HandleFunc("/notebooks", www.HandlerNotebooks)
	www.router.HandleFunc("/login", www.LoginFormHandler)
	www.router.HandleFunc("/personal", www.HandlerUserLogin)
	www.router.HandleFunc("/start-notebook", www.HandlerStartContainer)
	www.router.HandleFunc("/logout", www.LogutHandler)
	http.ListenAndServe(":8080", www.router)

}
