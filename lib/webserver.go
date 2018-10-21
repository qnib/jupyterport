package qniblib // import "github.com/qnib/jupyterport/lib"

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gorilla/mux"
	"github.com/kataras/go-sessions"
	"github.com/thedevsaddam/renderer"
	"github.com/urfave/negroni"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)


var (

	cookieNameForSessionID = "mycookiesessionnameid"
	tplDir = "./tpl/*.html"
)


type Webserver struct {
	rnd 			*renderer.Render
	sess   			*sessions.Sessions
	revProx			map[string]http.Handler
	SessionID		string
	router			*mux.Router
	database 		Database
	spawner 		Spawner
	jupyterImages	DockerImages
	notebookImages	DockerImages
	dataImages		DockerImages
	ctx         	*cli.Context
}

func NewWebserver(ctx *cli.Context) Webserver {
	return Webserver{
		ctx: ctx,
	}
}



func (www *Webserver) HandlerNotebooks(w http.ResponseWriter, r *http.Request) {
	var err error
	// Check if user is authenticated
	sess := www.sess.Start(w, r)
	cont := NewContent(sess.GetAll())
	cont.JupyterImages = www.jupyterImages.GetImages()
	cont.Notebooks, err = www.ListNotebooks(cont.User)
	cont.NotebookImages = www.notebookImages.GetImages()
	cont.DataImages = www.dataImages.GetImages()
	if err != nil {
		log.Println(err.Error())
		cont.Notebooks = make(map[string]Notebook)
		www.rnd.HTML(w, http.StatusOK,  "notebooks", cont)
		return
	}
	log.Printf("Content: %s", cont.String())
	if ! cont.Authenticated {
		http.Redirect(w, r, "/login", 303)
		return
	}
	www.rnd.HTML(w, http.StatusOK,  "notebooks", cont)
}

func (www *Webserver) ListNotebooks(user string) (nbs map[string]Notebook, err error) {
	return www.spawner.ListNotebooks(user, www.ctx.String("external-address"))
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

func (www *Webserver) HandlerStartContainer(w http.ResponseWriter, r *http.Request) {
	sess := www.sess.Start(w, r)
	nb, err := www.spawner.SpawnNotebook(sess.GetString("uname"), r, token, www.ctx.String("ext-addr"))
	cont := NewContent(sess.GetAll())
	www.rnd.HTML(w, http.StatusOK, "home", cont)
	log.Printf("Add route for user %s", cont.User)
	err = www.AddRoute(cont.User, r.FormValue("cntname"), nb.InternalUrl)
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

func (www *Webserver) Init(spawner Spawner, db Database) {
	opts := renderer.Options{
		ParseGlobPattern: tplDir,
	}
	di := []DockerImage{}
	for _, image := range www.ctx.StringSlice("jupyter-images") {
		log.Printf("Add jupyter-image: %s", image)
		di = append(di, DockerImage{Name: image})
	}
	www.jupyterImages = DockerImages{di}
	ni := []DockerImage{}
	for _, image := range www.ctx.StringSlice("notebook-images") {
		log.Printf("Add notebook-image: %s", image)
		ni = append(ni, DockerImage{Name: image})
	}
	www.notebookImages = DockerImages{ni}
	dataI := []DockerImage{}
	for _, image := range www.ctx.StringSlice("data-images") {
		log.Printf("Add data-image: %s", image)
		dataI = append(dataI, DockerImage{Name: image})
	}
	www.dataImages = DockerImages{dataI}
	www.database = db
	spawner.Init(www.ctx)
	www.spawner = spawner
	www.router = mux.NewRouter()
	www.rnd = renderer.New(opts)
	www.sess = sessions.New(sessions.Config{Cookie: cookieNameForSessionID})

}

func (www *Webserver) AddRoute(uid, cntname, target string) (err error) {
	remote, err := url.Parse(target)
	if err != nil {
		return
	}
	prxy := httputil.NewSingleHostReverseProxy(remote)
	link := fmt.Sprintf("/user/%s/%s.*", uid, cntname)
	log.Printf("%s -> %s", link, target)
	www.router.HandleFunc(link, handler(prxy ,target)).Methods("GET", "PUT", "HEAD", "OPTIONS")
	return
}

func handler(p *httputil.ReverseProxy, targetBase string) func(http.ResponseWriter, *http.Request) {
	// TODO: Use this as a function for `/user/` and match the targeted notebook dynamically.
	//
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Proxy > r.URL.Path:%s // r.URL.RawQuery: %v // Connection:%s // Upgrade:%v", r.URL.Path, r.URL.RawQuery, r.Header["Connection"], r.Header["Upgrade"])
			if !IsWebSocket(r) {
				p.ServeHTTP(w, r)
			} else {
				target := fmt.Sprintf("%s%s", targetBase, r.URL.Path)
				log.Println("WebSocket-target:", target)
				dialer := net.Dialer{KeepAlive: time.Second * 10}
				d, err := dialer.Dial("tcp", target)
				if err != nil {
					log.Printf("ERROR: dialing websocket backend '%s': %v\n", target, err)
					http.Error(w, "Error contacting backend server.", 500)
					return
				}
				hj, ok := w.(http.Hijacker)
				if !ok {
					log.Println("ERROR: Not Hijackable")
					http.Error(w, "Internal Error: Not Hijackable", 500)
					return
					return
				}
				nc, _, err := hj.Hijack()
				if err != nil {
					log.Printf("ERROR: Hijack error: %v\n", err)
					return
				}
				defer nc.Close()
				defer d.Close()

				// copy the request to the target first
				err = r.Write(d)
				if err != nil {
					log.Printf("ERROR: copying request to target: %v\n", err)
					return
				}

				errc := make(chan error, 2)
				cp := func(dst io.Writer, src io.Reader) {
					_, err := io.Copy(dst, src)
					errc <- err
				}
				go cp(d, nc)
				go cp(nc, d)
				<-errc
		}
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
	// TODO: make it dynamic
	target := "test-mynotebook.default.svc.cluster.local:8888"
	remote, _ := url.Parse(fmt.Sprintf("http://%s",target))
	prxy := httputil.NewSingleHostReverseProxy(remote)
	www.router.HandleFunc("/user/test/mynotebook/{rest:.*}", handler(prxy, target))
	addr := www.ctx.String("listen-addr")
	log.Printf("Start ListenAndServe on address '%s'", addr)
	n := negroni.New(negroni.NewLogger())
	// Or use a middleware with the Use() function
	n.UseHandler(www.router)
	http.ListenAndServe(addr, n)

}
