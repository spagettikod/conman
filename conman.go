package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func listContainers(w http.ResponseWriter, r *http.Request) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// arg := filters.NewArgs()
	// arg.Add("label", "uid=rolbal")
	cs, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}

	containers := []container{}
	for _, c := range cs {
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

func downloadLog(w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)
	containerID := vars["id"]
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

func wrapper(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func main() {
	r := mux.NewRouter()

	s := r.PathPrefix("/api").Subrouter()
	s.HandleFunc("/containers", wrapper(listContainers))
	s.HandleFunc("/containers/{id}/log/download", wrapper(downloadLog)).Methods("GET")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("/www")))
	http.ListenAndServe(":80", r)
}
