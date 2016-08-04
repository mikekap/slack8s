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
	count     int
	eventType string
}

type messager interface {
	sendMessage(message) error
}

func getWhitelist(wlStr string) (wl whitelist) {
	if wlStr == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(wlStr), &wl); err != nil {
		return nil
	}

	return wl
}

func getTypes(typeStr string) (tm map[watch.EventType]bool) {
	if typeStr == "" {
		return nil
	}

	var types []watch.EventType
	if err := json.Unmarshal([]byte(typeStr), &types); err != nil {
		return nil
	}

	tm = make(map[watch.EventType]bool)
	for _, t := range types {
		tm[t] = true
	}

	return tm
}

func main() {
	cl, err := k8s.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	kube := &kubeCfg{
		kubeClient: cl,
		types:      getTypes(os.Getenv("EVENT_TYPES")),
		whitelist:  getWhitelist(os.Getenv("WHITELIST")),
	}

	if kube.types != nil {
		log.Printf("types filtered: %v", kube.types)
	}

	if kube.whitelist != nil {
		log.Printf("whitelist configured:")
		for _, w := range kube.whitelist {
			log.Printf("entry: %+v", w)
		}
	}

	msgr := &slackCfg{
		messagePoster: slack.New(os.Getenv("SLACK_TOKEN")),
		channel:       os.Getenv("SLACK_CHANNEL"),
		env:           os.Getenv("ENV"),
	}

	log.Printf("posting to channel %s", msgr.channel)
	if msgr.env != "" {
		log.Printf("running in env %s", msgr.env)
	}

	err = kube.watchEvents(msgr)
	if err != nil {
		log.Fatalf("Error from watchEvents(): %s", err)
	}

	log.Println("Terminating...")
}
