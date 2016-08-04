package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/nlopes/slack"

	k8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type message struct {
	msg       string
	obj       string
	name      string
	reason    string
	component string
	color     string
}

type messager interface {
	sendMessage(message) error
}

func main() {
	cl, err := k8s.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	kube := &kubeCfg{
		kubeClient: cl,
	}

	msgr := &slackCfg{
		messagePoster: slack.New(os.Getenv("SLACK_TOKEN")),
		channel:       os.Getenv("SLACK_CHANNEL"),
		env:           os.Getenv("ENV"),
	}

	var types []watch.EventType
	if err = json.Unmarshal([]byte(os.Getenv("EVENT_TYPES")), &types); err != nil {
		typeMap := make(map[watch.EventType]bool)

		for _, t := range types {
			typeMap[t] = true
		}

		kube.types = typeMap
	}

	err = kube.watchEvents(msgr)
	if err != nil {
		log.Fatalf("Error from watchEvents(): %s", err)
	}

	log.Println("Terminating...")
}
