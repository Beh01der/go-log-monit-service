package main

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"flag"
)


func main() {
	filterLabel := flag.String("label", "monitor-logs", "Look for this label in docker containers")
	flag.Parse()
	fmt.Printf("Looking for containers labeled with: %s\n", *filterLabel)

	startDockerMonitor(DockerMonitorConfig{filterLabel: *filterLabel, handler: func(container docker.APIContainers, up bool){
		if up {
			fmt.Printf("Container Up: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
		} else {
			fmt.Printf("Container Down: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
		}
	}})

	fmt.Println("Terminating the service")
}

