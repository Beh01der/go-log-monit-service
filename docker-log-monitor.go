package main

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/hpcloud/tail"
	"time"
	"fmt"
	"os"
	"encoding/json"
)

type DockerLogMonitorConfig struct {
	Endpoint    string
	FilterLabel string
	ReadFromStart bool
	Handler     func(container *docker.APIContainers, logEntry *LogEntry)
}

type LogEntry struct {
	Log string `json:"log"`
	Stream string `json:"stream"`
	Time time.Time `json:"time"`
}

var watchers map[string]*tail.Tail

func StartDockerLogMonitor(config DockerLogMonitorConfig) {
	if config.FilterLabel == "" {
		config.FilterLabel = "monitor-logs"
	}

	watchers = make(map[string]*tail.Tail)

	StartDockerMonitor(DockerMonitorConfig{FilterLabel: config.FilterLabel, Handler: func(container docker.APIContainers, up bool) {
		if up {
			fmt.Printf("Container Up: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
			logFile := "/var/lib/docker/containers/" + container.ID + "/" + container.ID + "-json.log"
			fileInfo, err := os.OpenFile(logFile, os.O_RDONLY, 0)
			if err == nil {
				fileStat, err := fileInfo.Stat()
				if err == nil {
					fileOffset := int64(0)
					if !config.ReadFromStart {
						fileOffset = fileStat.Size()
					}

					watcher, err := tail.TailFile(logFile, tail.Config{Follow: true, Location: &tail.SeekInfo{ Offset: fileOffset}})
					if err == nil {
						watchers[container.ID] = watcher
						go func() {
							fmt.Printf("Watching log file for service %s offset %d : %s\n", container.Names, fileOffset, logFile)
							logEntry := LogEntry{}
							for line := range watcher.Lines {
								err := json.Unmarshal([]byte(line.Text), &logEntry)
								if err == nil {
									if config.Handler != nil {
										config.Handler(&container, &logEntry)
									}
								} else {
									fmt.Printf("Error parsing new line %e\n", err)
								}
							}
							watcher.Cleanup()
							fmt.Printf("Stopped watching log file for service %s\n", container.Names)
						}()
					}
				}
			}
		} else {
			fmt.Printf("Container Down: ID=%s Labels=%s Image=%s Names=%s\n", container.ID, container.Labels, container.Image, container.Names)
			if watcher, ok := watchers[container.ID]; ok {
				watcher.Stop()
			}
		}
	}})
}