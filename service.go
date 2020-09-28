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
)

type ServiceLinks struct {
	DownloadLog *hateoasLink `json:"downloadLog,omitempty"`
}

type Service struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Image string       `json:"image"`
	Links ServiceLinks `json:"links"`
}

func NewDownloadServiceLogLink(id string) *hateoasLink {
	return &hateoasLink{Href: fmt.Sprintf("/api/services/%s/log/download", id), Rel: "downloadLog", Type: "GET"}
}

func ListServices(auth Authenticator) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		cli, err := client.NewEnvClient()
		if err != nil {
			return err
		}

		serviceList, err := cli.ServiceList(context.Background(), types.ServiceListOptions{})
		if err != nil {
			return err
		}

		services := []Service{}
		for _, svc := range serviceList {
			service := Service{}
			allowed, err := auth.IsServiceAllowed(r, svc.ID)
			if err != nil {
				return err
			}
			if !allowed {
				continue
			}
			service.ID = svc.ID
			service.Name = svc.Spec.Name
			service.Image = svc.Spec.TaskTemplate.ContainerSpec.Image
			service.Links.DownloadLog = NewDownloadServiceLogLink(svc.ID)
			services = append(services, service)
		}
		b, err := json.Marshal(services)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return err
		}
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, string(b))
		return nil
	}
}

func DownloadServiceLog(serviceID string, w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, _ := client.NewEnvClient()
	cjson, err := client.ContainerInspect(context.Background(), serviceID)
	if err != nil {
		return err
	}

	w.Header().Set("Content-type", "text/plain;charset=UTF-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+cjson.Name+"\"")

	creader, err := client.ContainerLogs(ctx, serviceID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: false, Details: false, Timestamps: false})
	if err != nil {
		return err
	}

	_, err = io.Copy(w, creader)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}
