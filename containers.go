package main

import (
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Docker struct {
	cli     *client.Client
	ctx     context.Context
	filters filters.Args
}

func NewDocker(contFilter string) *Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	filterMap := filters.NewArgs()
	return &Docker{
		cli:     cli,
		ctx:     context.Background(),
		filters: extractFilters(contFilter, filterMap),
	}
}

func extractFilters(filter string, filterMap filters.Args) filters.Args {

	if filter != "" {
		argStr := strings.Split(filter, ",")
		for _, v := range argStr {
			bef, aft, found := strings.Cut(v, "=")
			if !found {
				log.Printf("Unable to find = in string %v. Thus skipping it\n", v)
				continue
			}
			filterMap.Add(bef, aft)
		}
	}
	log.Printf("Final Filter Arguments to use are: %v\n", filterMap)
	return filterMap
}

func (d *Docker) SearchContainers(limit int, contFilter string) ([]types.Container, error) {

	filterMap := d.filters.Clone()
	if contFilter != "" {
		filterMap = extractFilters(contFilter, filterMap)
	}

	containerOptions := container.ListOptions{Limit: limit, Filters: filterMap}
	containerList, err := d.cli.ContainerList(d.ctx, containerOptions)
	if err != nil {
		log.Printf("Unable to find container with these conditions due to: %v", err)
	}

	return containerList, err
}

func (d *Docker) StartContainer(container_id string) error {

	err := d.cli.ContainerStart(d.ctx, container_id, container.StartOptions{})
	if err != nil {
		log.Printf("Unable to start container %v", container_id)
	}
	return err
}

func (d *Docker) StopContainer(container_id string) error {

	err := d.cli.ContainerStop(d.ctx, container_id, container.StopOptions{})
	if err != nil {
		log.Printf("Unable to start container %v", container_id)
	}
	return err
}

func (d *Docker) StatusContainer(container_id string) string {

	var conStatus string
	filterMap := filters.NewArgs()
	filterMap.Add("id", container_id)
	options := container.ListOptions{Limit: 1, Filters: filterMap}
	contList, err := d.cli.ContainerList(d.ctx, options)
	if err != nil {
		log.Printf("Unable to start container %v", container_id)
		conStatus = "Error occured"
	}
	if len(contList) == 0 {
		log.Printf("Unable to find container with id: %v", container_id)
		conStatus = "Error occured"
	}
	conStatus = contList[0].Status
	return conStatus
}

func (d *Docker) RestartContainer(container_id string) error {

	err := d.StopContainer(container_id)
	if err != nil {
		log.Printf("Unable to start container %v", container_id)
	}
	err = d.StartContainer(container_id)
	if err != nil {
		log.Printf("Unable to start container %v", container_id)
	}
	return err
}

func (d *Docker) Close() {
	d.cli.Close()
}
