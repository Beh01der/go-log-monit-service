package main

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"flag"
)

type TrackedEventType int

const (
	UNKNOWN_EVENT TrackedEventType = iota
	POSITIVE_EVENT
	NEGATIVE_EVENT
)

const endpoint = "unix:///var/run/docker.sock"

var falseStrings map[string]bool
var trackedEvents map[string]TrackedEventType
var containerById map[string]docker.APIContainers
var filterLabel *string

func main() {
	filterLabel = flag.String("label", "monitor-logs", "Look for this label in docker containers")
	flag.Parse()
	fmt.Printf("Looking for containers labeled with: %s\n", *filterLabel)

	falseStrings = map[string]bool{
		"0": true,
		"null": true,
		"false": true,
		"disable": true,
		"disabled": true,
		"": true,
	}

	trackedEvents = map[string]TrackedEventType{
		"create": POSITIVE_EVENT,
		"restart": POSITIVE_EVENT,
		"start": POSITIVE_EVENT,
		"destroy": NEGATIVE_EVENT,
		"die": NEGATIVE_EVENT,
		"kill": NEGATIVE_EVENT,
		"stop": NEGATIVE_EVENT,
	}

	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	containerById = make(map[string]docker.APIContainers)

	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		updateContainer(container)
	}

	// listen for events
	eventListener := make(chan *docker.APIEvents)
	client.AddEventListener(eventListener)

	for {
		event := <-eventListener
		trackedEvent := trackedEvents[event.Status]
		if trackedEvent == POSITIVE_EVENT {
			if _, ok := containerById[event.ID]; !ok {
				// if new container
				containers, err := client.ListContainers(docker.ListContainersOptions{})
				if err != nil {
					panic(err)
				}

				for _, container := range containers {
					if container.ID == event.ID {
						updateContainer(container)
						break
					}
				}
			}
		} else if trackedEvent == NEGATIVE_EVENT {
			if container, ok := containerById[event.ID]; ok {
				removeContainer(container)
			}
		}
	}

	fmt.Println("Terminating the service")
}

func updateContainer(container docker.APIContainers) {
	if container.Labels != nil && !falseStrings[container.Labels[*filterLabel]] {
		fmt.Printf("Container Up: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
		containerById[container.ID] = container
	}
}

func removeContainer(container docker.APIContainers) {
	fmt.Printf("Container Down: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
	delete(containerById, container.ID)
}

