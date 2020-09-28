package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

type hateoasLink struct {
	Href string `json:"href,omitempty"`
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
}

func errLogWrapper(errLog, auditLog *log.Logger, fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		auditLog.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL, r.UserAgent())
		err := fn(w, r)
		if err != nil {
			errLog.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func authContainerWrapper(auth Authenticator, fn func(containerID string, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		containerID := vars["id"]
		allowed, err := auth.IsContainerAllowed(r, containerID)
		if err != nil {
			return err
		}
		if !allowed {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return fn(containerID, w, r)
	}
}

func authServiceWrapper(auth Authenticator, fn func(serviceID string, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		serviceID := vars["id"]
		allowed, err := auth.IsServiceAllowed(r, serviceID)
		if err != nil {
			return err
		}
		if !allowed {
			w.WriteHeader(http.StatusForbidden)
			return nil
		}
		return fn(serviceID, w, r)
	}
}

func main() {
	router := mux.NewRouter()

	var auth Authenticator
	if os.Getenv("CONMAN_AUTH") == "HTTP" {
		header := os.Getenv("CONMAN_AUTH_HTTP_HEADER")
		if header == "" {
			log.Fatalln("Environment variable CONMAN_AUTH is set to HTTP but the variable CONMAN_AUTH_HTTP_HEADER is not set")
		}
		auth = HTTPHeaderAuthenticator{HTTPHeader: header, ContainerLabelKey: "conman.auth.id"}
	} else {
		auth = NoOpAuthenticator{}
	}

	urlRoot, rootSet := os.LookupEnv("CONMAN_URL_ROOT")
	if !rootSet {
		urlRoot = ""
	}

	auditLog := log.New(ioutil.Discard, "AUDIT ", log.LstdFlags)
	_, auditEnv := os.LookupEnv("CONMAN_LOG_AUDIT")
	if auditEnv {
		auditLog.SetOutput(os.Stdout)
	}
	errLog := log.New(os.Stdout, "ERROR ", log.LstdFlags)

	apiRouter := router.PathPrefix(urlRoot + "/api").Subrouter()
	apiRouter.HandleFunc("/containers", errLogWrapper(errLog, auditLog, ListContainers(auth)))
	apiRouter.HandleFunc("/containers/{id}/log/download", errLogWrapper(errLog, auditLog, authContainerWrapper(auth, DownloadContainerLog))).Methods("GET")
	apiRouter.HandleFunc("/containers/{id}", errLogWrapper(errLog, auditLog, authContainerWrapper(auth, RemoveContainer))).Methods("DELETE")
	apiRouter.HandleFunc("/services", errLogWrapper(errLog, auditLog, ListServices(auth)))
	apiRouter.HandleFunc("/services/{id}/log/download", errLogWrapper(errLog, auditLog, authServiceWrapper(auth, DownloadServiceLog))).Methods("GET")

	router.PathPrefix(urlRoot + "/").Handler(http.StripPrefix(urlRoot, http.FileServer(http.Dir("/www"))))
	router.PathPrefix(urlRoot).Handler(http.RedirectHandler(urlRoot+"/", http.StatusMovedPermanently))
	http.ListenAndServe(":8080", router)
}
