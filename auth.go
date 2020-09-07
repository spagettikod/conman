package main

import (
	"context"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Authenticator interface {
	IsContainerAllowed(r *http.Request, containerID string) (bool, error)
	IsServiceAllowed(r *http.Request, serviceID string) (bool, error)
}

type NoOpAuthenticator struct {
}

func (noa NoOpAuthenticator) IsContainerAllowed(r *http.Request, containerID string) (bool, error) {
	return true, nil
}

func (noa NoOpAuthenticator) IsServiceAllowed(r *http.Request, serviceID string) (bool, error) {
	return true, nil
}

type HTTPHeaderAuthenticator struct {
	HTTPHeader        string
	ContainerLabelKey string
}

func (hha HTTPHeaderAuthenticator) IsContainerAllowed(r *http.Request, containerID string) (bool, error) {
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

func (hha HTTPHeaderAuthenticator) IsServiceAllowed(r *http.Request, serviceID string) (bool, error) {
	if len(r.Header[hha.HTTPHeader]) < 1 {
		return false, nil
	}
	subject := r.Header[hha.HTTPHeader][0]
	labelValue, err := getServiceLabel(hha.ContainerLabelKey, serviceID)
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

func getServiceLabel(labelKey, serviceID string) (labelValue string, err error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return
	}

	svc, _, err := cli.ServiceInspectWithRaw(context.Background(), serviceID)
	if err != nil {
		return
	}
	return svc.Spec.Labels[labelKey], err
}
