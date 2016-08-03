package main

import (
	"log"
	"os"

	"github.com/nlopes/slack"

	k8s "k8s.io/kubernetes/pkg/client/unversioned"
)

type message struct {
	msg       string
	obj       string
	name      string
	reason    string
	component string
	color     string
	count     int
}

type messager interface {
	sendMessage(message) error
}

func main() {
	cl, err := k8s.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	msgr := &slackCfg{
		messagePoster: slack.New(os.Getenv("SLACK_TOKEN")),
		channel:       os.Getenv("SLACK_CHANNEL"),
	}

	err = watchEvents(cl, msgr)
	if err != nil {
		log.Fatalf("Error from watchEvents(): %s", err)
	}

	log.Println("Terminating...")
}
