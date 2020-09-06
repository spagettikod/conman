package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
)

type hateoasLink struct {
	Href string `json:"href,omitempty"`
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
}

type containerLinks struct {
	DownloadLog *hateoasLink `json:"downloadLog,omitempty"`
}

func newDownloadLogLink(id string) *hateoasLink {
	return &hateoasLink{Href: fmt.Sprintf("/api/containers/%s/log/download", id), Rel: "downloadLog", Type: "GET"}
}

type container struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Image  string         `json:"image"`
	Status string         `json:"status"`
	State  string         `json:"state"`
	Links  containerLinks `json:"links"`
}

type Authenticator interface {
	IsAllowed(r *http.Request, containerID string) (bool, error)
}

type NoOpAuthenticator struct {
}

func (noa NoOpAuthenticator) IsAllowed(r *http.Request, containerID string) (bool, error) {
	return true, nil
}

type HTTPHeaderAuthenticator struct {
	HTTPHeader        string
	ContainerLabelKey string
}

func (hha HTTPHeaderAuthenticator) IsAllowed(r *http.Request, containerID string) (bool, error) {
	if len(r.Header[hha.HTTPHeader]) < 1 {
		return false, nil
	}
	subject := r.Header[hha.HTTPHeader][0]
	labelValue, err := getContainerLabel(hha.ContainerLabelKey, containerID)
	if err != nil {
		return false, err
	}
	return (subject == labelValue), nil
}

func getContainerLabel(labelKey, containerID string) (labelValue string, err error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return
	}
	// TODO Is there no other way to get the labels, do I need to list?
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return
	}
	for _, container := range containers {
		if container.ID == containerID {
			labelValue = container.Labels[labelKey]
			return
		}
	}
	return
}

func listContainers(auth Authenticator) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		cs, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			return err
		}

		containers := []container{}
		for _, c := range cs {
			allowed, err := auth.IsAllowed(r, c.ID)
			if err != nil {
				return err
			}
			if !allowed {
				continue
			}
			ci, err := cli.ContainerInspect(context.Background(), c.ID)
			if err != nil {
				return err
			}
			image, _, err := cli.ImageInspectWithRaw(context.Background(), ci.Image)
			if err != nil {
				return err
			}
			container := container{ID: ci.ID, Name: ci.Name[1:], State: c.State, Status: c.Status}
			if len(image.RepoTags) > 0 {
				container.Image = image.RepoTags[0]
			}
			container.Links.DownloadLog = newDownloadLogLink(container.ID)
			containers = append(containers, container)
		}
		b, err := json.Marshal(containers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, string(b))
		return nil
	}
}

func downloadLog(containerID string, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, _ := client.NewEnvClient()
	cjson, err := client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/plain;charset=UTF-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+cjson.Name+"\"")

	creader, err := client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: false, Details: false, Timestamps: false})
	if err != nil {
		return err
	}

	_, err = io.Copy(w, creader)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func errLogWrapper(auditLog *log.Logger, fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		auditLog.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL, r.UserAgent())
		err := fn(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func authWrapper(auth Authenticator, fn func(containerID string, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		containerID := vars["id"]
		allowed, err := auth.IsAllowed(r, containerID)
		if err != nil {
			return err
		}
		if !allowed {
			w.WriteHeader(http.StatusUnauthorized)
			return nil
		}
		return fn(containerID, w, r)
	}
}

func main() {
	r := mux.NewRouter()

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

	auditLog := log.New(os.Stdout, "AUDIT", log.LstdFlags)

	s := r.PathPrefix("/api").Subrouter()
	s.HandleFunc("/containers", errLogWrapper(auditLog, listContainers(auth)))
	s.HandleFunc("/containers/{id}/log/download", errLogWrapper(auditLog, authWrapper(auth, downloadLog))).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/www")))
	http.ListenAndServe(":80", r)
}
