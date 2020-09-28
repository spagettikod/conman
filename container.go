package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type ContainerLinks struct {
	DownloadLog *hateoasLink `json:"downloadLog,omitempty"`
	Remove      *hateoasLink `json:"remove,omitempty"`
}

func NewDownloadContainerLogLink(id string) *hateoasLink {
	return &hateoasLink{Href: fmt.Sprintf("/api/containers/%s/log/download", id), Rel: "downloadLog", Type: "GET"}
}

func NewRemoveContainerLink(id string) *hateoasLink {
	return &hateoasLink{Href: fmt.Sprintf("/api/containers/%s", id), Rel: "remove", Type: "DELETE"}
}

type Container struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Image  string         `json:"image"`
	Status string         `json:"status"`
	State  string         `json:"state"`
	Links  ContainerLinks `json:"links"`
}

func ListContainers(auth Authenticator) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		cs, err := cli.ContainerList(context.Background(), types.ContainerListOptions{All: true})
		if err != nil {
			return err
		}

		containers := []Container{}
		for _, c := range cs {
			allowed, err := auth.IsContainerAllowed(r, c.ID)
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
			container := Container{ID: ci.ID, Name: ci.Name[1:], State: c.State, Status: c.Status}
			if len(image.RepoTags) > 0 {
				container.Image = image.RepoTags[0]
			}
			switch c.State {
			case "exited":
				container.Links.Remove = NewRemoveContainerLink(container.ID)
			}
			container.Links.DownloadLog = NewDownloadContainerLogLink(container.ID)
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

func RemoveContainer(containerID string, w http.ResponseWriter, r *http.Request) error {
	client, _ := client.NewEnvClient()
	err := client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func StopContainer(containerID string, w http.ResponseWriter, r *http.Request) error {
	client, _ := client.NewEnvClient()
	d := 5 * time.Second
	err := client.ContainerStop(context.Background(), containerID, &d)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func DownloadContainerLog(containerID string, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, _ := client.NewEnvClient()
	cjson, err := client.ContainerInspect(context.Background(), containerID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/plain;charset=UTF-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+cjson.Name+"\"")

	reader, err := client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: false, Details: false, Timestamps: false})
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()[8:] + "\n"
		w.Write([]byte(line))
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
