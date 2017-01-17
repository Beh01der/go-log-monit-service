package main

import (
	"fmt"
	"flag"
	"github.com/fsouza/go-dockerclient"
	"github.com/vjeantet/grok"
	"github.com/cactus/go-statsd-client/statsd"
	"strings"
	"strconv"
)

const logPattern = `%{IP:client} \[%{TIMESTAMP_ISO8601:timestamp}\] "%{WORD:method} %{URIHOST:site}%{URIPATHPARAM:url}" %{INT:code} %{INT:request} %{INT:response} - %{NUMBER:took} \[%{DATA:cache}\] "%{DATA:mtag}" "%{DATA:agent}"`

func main() {
	filterLabel := flag.String("label", "monitor-logs", "Look for this label in docker containers")
	statsdAddr := flag.String("statsd", "127.0.0.1:8125", "StatsD address to send stats to")

	flag.Parse()
	fmt.Printf("Looking for containers labeled with: %s\n", *filterLabel)
	fmt.Printf("Using StatsD on: %s\n", *statsdAddr)

	sdc, err := statsd.NewClient(*statsdAddr, "statsd-client")
	if err != nil {
		panic(err)
	}
	defer sdc.Close()

	g, _ := grok.New()

	StartDockerLogMonitor(DockerLogMonitorConfig{FilterLabel:*filterLabel, Handler:func(container *docker.APIContainers, logEntry *LogEntry) {
		values, _ := g.Parse(logPattern, logEntry.Log)

		code, _ := values["code"]
		url, _ := values["url"]

		if codeInt, _ := strconv.Atoi(code); codeInt > 99 {
			sdc.Inc("router.hit", 1, 1.0)
			sdc.Inc("router.hit." + code, 1, 1.0)
		}

		if url != "" {
			if strings.Contains(url, "api/note") {
				sdc.Inc("api.note.hit", 1, 1.0)
				sdc.Inc("api.note", 1, 1.0)
			} else if strings.Contains(url, "api/policy") {
				sdc.Inc("api.hit", 1, 1.0)
			}
		}
	}})

	fmt.Println("Terminating the service")
}
