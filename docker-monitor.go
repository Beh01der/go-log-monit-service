package main

import "github.com/fsouza/go-dockerclient"

type TrackedEventType int

const (
	UNKNOWN_EVENT TrackedEventType = iota
	POSITIVE_EVENT
	NEGATIVE_EVENT
)

type DockerMonitorConfig struct {
	Endpoint    string
	FilterLabel string
	Handler     func(container docker.APIContainers, up bool)
}

var falseStrings map[string]bool
var trackedEvents map[string]TrackedEventType
var containerById map[string]docker.APIContainers

func StartDockerMonitor(config DockerMonitorConfig) {
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

	if config.Endpoint == "" {
		config.Endpoint = "unix:///var/run/docker.sock"
	}

	if config.FilterLabel == "" {
		config.FilterLabel = "monitor-container"
	}

	dockerClient, err := docker.NewClient(config.Endpoint)
	if err != nil {
		panic(err)
	}

	containerById = make(map[string]docker.APIContainers)

	containers, err := dockerClient.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		if container.Labels != nil && !falseStrings[container.Labels[config.FilterLabel]] {
			config.updateContainer(container)
		}
	}

	// listen for events
	eventListener := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventListener)

	for {
		event := <-eventListener
		if trackedEvent := trackedEvents[event.Status]; trackedEvent == POSITIVE_EVENT {
			if _, ok := containerById[event.ID]; !ok {
				// if new container
				containers, err := dockerClient.ListContainers(docker.ListContainersOptions{})
				if err != nil {
					panic(err)
				}

				for _, container := range containers {
					if container.ID == event.ID {
						if container.Labels != nil && !falseStrings[container.Labels[config.FilterLabel]] {
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
	if config.Handler != nil {
		config.Handler(container, true)
	}
}

func (config *DockerMonitorConfig) removeContainer(container docker.APIContainers) {
	delete(containerById, container.ID)
	if config.Handler != nil {
		config.Handler(container, false)
	}
}

