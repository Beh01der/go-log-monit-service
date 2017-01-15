package main

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
)

func main()  {

	endpoint := "unix:///var/run/docker.sock"
	client, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	eventListener := make(chan *docker.APIEvents)
	client.AddEventListener(eventListener)

	for {
		event := <-eventListener
		fmt.Printf("ID=%s, Status=%s, Type=%s, Action=%s, From=%s, Actor=%s", event.ID, event.Status, event.Type, event.Action, event.From, event.Actor)
	}

	fmt.Println("Terminating the service")
}
