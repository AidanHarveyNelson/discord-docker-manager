package containers

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Docker struct {
	cli *client.Client
	ctx context.Context
}

func NewDocker() *Docker {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Docker{
		cli: cli,
		ctx: context.Background(),
	}
}

func (d *Docker) SearchContainers(limit int) ([]types.Container, error) {

	filterMap := filters.NewArgs()
	// To-Do unpack these values properly so that multiple filters can be supported
	filterMap.Add("label", os.Getenv("DOCKER_FILTER"))
	containerOptions := container.ListOptions{Limit: limit, Filters: filterMap}
	containerList, err := d.cli.ContainerList(d.ctx, containerOptions)
	if err != nil {
		log.Printf("Unable to find container with these conditions. %v", err)
	}

	return containerList, err
}

func (d *Docker) StartContainer(container_id string) error {

	err := d.cli.ContainerStart(d.ctx, container_id, container.StartOptions{})
	if err != nil {
		fmt.Printf("Unable to start container %v", container_id)
	}
	return err
}

func (d *Docker) StopContainer(container_id string) error {

	err := d.cli.ContainerStop(d.ctx, container_id, container.StopOptions{})
	if err != nil {
		fmt.Printf("Unable to start container %v", container_id)
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
		fmt.Printf("Unable to start container %v", container_id)
		conStatus = "Error occured"
	}
	if len(contList) == 0 {
		fmt.Printf("Unable to find container with id: %v", container_id)
		conStatus = "Error occured"
	}
	conStatus = contList[0].Status
	return conStatus
}

func (d *Docker) RestartContainer(container_id string) error {

	err := d.StopContainer(container_id)
	if err != nil {
		fmt.Printf("Unable to start container %v", container_id)
	}
	err = d.StartContainer(container_id)
	if err != nil {
		fmt.Printf("Unable to start container %v", container_id)
	}
	return err
}
