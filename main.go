package main

import (
	"fmt"
	"flag"
	"github.com/fsouza/go-dockerclient"
)

func main() {
	filterLabel := flag.String("label", "monitor-logs", "Look for this label in docker containers")
	flag.Parse()
	fmt.Printf("Looking for containers labeled with: %s\n", *filterLabel)

	StartDockerLogMonitor(DockerLogMonitorConfig{FilterLabel:*filterLabel, Handler:func(container *docker.APIContainers, logEntry *LogEntry) {
		fmt.Printf("New log entry %v\n", logEntry)
	}})

	fmt.Println("Terminating the service")
}

