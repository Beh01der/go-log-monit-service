package main

import (
	"fmt"
	"flag"
	"github.com/fsouza/go-dockerclient"
	"github.com/vjeantet/grok"
	"strings"
	"strconv"
	"github.com/quipo/statsd"
	"time"
)

const logPattern = `%{IP:client} \[%{TIMESTAMP_ISO8601:timestamp}\] "%{WORD:method} %{URIHOST:site}%{URIPATHPARAM:url}" %{INT:code} %{INT:request} %{INT:response} - %{NUMBER:took} \[%{DATA:cache}\] "%{DATA:mtag}" "%{DATA:agent}"`

func main() {
	filterLabel := flag.String("label", "monitor-logs", "Look for this label in docker containers")
	statsdAddr := flag.String("statsd", "127.0.0.1:8125", "StatsD address to send stats to")

	flag.Parse()
	fmt.Printf("Looking for containers labeled with: %s\n", *filterLabel)
	fmt.Printf("Using StatsD on: %s\n", *statsdAddr)

	sdc := statsd.NewStatsdClient(*statsdAddr, "")
	err := sdc.CreateSocket()
	if err != nil {
		panic(err)
	}
	stats := statsd.NewStatsdBuffer(time.Second * 2, sdc)
	defer sdc.Close()

	g, _ := grok.New()

	StartDockerLogMonitor(DockerLogMonitorConfig{FilterLabel:*filterLabel, Handler:func(container *docker.APIContainers, logEntry *LogEntry) {
		values, err := g.Parse(logPattern, logEntry.Log)
		if err != nil {
			return
		}

		code, _ := values["code"]
		url, _ := values["url"]

		if codeInt, err := strconv.Atoi(code); err == nil && codeInt > 99 {
			stats.Incr("router.hit", 1)
			stats.Incr("router.hit." + code, 1)
		}

		if url != "" {
			if strings.Contains(url, "api/note") {
				stats.Incr("api.note.hit", 1)
				stats.Incr("api.note", 1)
			} else if strings.Contains(url, "api/policy") {
				stats.Incr("api.hit", 1)
			}
		}
	}})

	fmt.Println("Terminating the service")
}
