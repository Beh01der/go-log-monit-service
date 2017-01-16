package main

import "github.com/fsouza/go-dockerclient"

type TrackedEventType int

const (
	UNKNOWN_EVENT TrackedEventType = iota
	POSITIVE_EVENT
	NEGATIVE_EVENT
)

type DockerMonitorConfig struct {
	endpoint    string
	filterLabel string
	handler     func(container docker.APIContainers, up bool)
}

var falseStrings map[string]bool
var trackedEvents map[string]TrackedEventType
var containerById map[string]docker.APIContainers

func startDockerMonitor(config DockerMonitorConfig) {
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

	endpoint := config.endpoint
	if endpoint == "" {
		endpoint = "unix:///var/run/docker.sock"
	}

	filterLabel := config.filterLabel
	if filterLabel == "" {
		filterLabel = "monitor-logs"
	}

	dockerClient, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	containerById = make(map[string]docker.APIContainers)

	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if container.Labels != nil && !falseStrings[container.Labels[filterLabel]] {
			config.updateContainer(container)
		}
	}

	// listen for events
	eventListener := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventListener)

	for {
		event := <-eventListener
		trackedEvent := trackedEvents[event.Status]
		if trackedEvent == POSITIVE_EVENT {
			if _, ok := containerById[event.ID]; !ok {
				// if new container
				containers, err := dockerClient.ListContainers(docker.ListContainersOptions{})
				if err != nil {
					panic(err)
				}

				for _, container := range containers {
					if container.ID == event.ID {
						if container.Labels != nil && !falseStrings[container.Labels[filterLabel]] {
							config.updateContainer(container)
						}
						break
					}
				}
			}
		} else if trackedEvent == NEGATIVE_EVENT {
			if container, ok := containerById[event.ID]; ok {
				config.removeContainer(container)
			}
		}
	}
}

func (config *DockerMonitorConfig) updateContainer(container docker.APIContainers) {
	containerById[container.ID] = container
	if config.handler != nil {
		config.handler(container, true)
	}
}

func (config *DockerMonitorConfig) removeContainer(container docker.APIContainers) {
	delete(containerById, container.ID)
	if config.handler != nil {
		config.handler(container, false)
	}
}

