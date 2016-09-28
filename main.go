package main

import (
	"encoding/json"
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

func main() {
	cl, err := k8s.NewInCluster()
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
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

	kube := &kubeCfg{
		Client:     cl,
		whitelist:  getWhitelist(os.Getenv("WHITELIST")),
		msgr:       msgr,
	}

	if kube.whitelist != nil {
		log.Printf("whitelist configured:")
		for _, w := range kube.whitelist {
			log.Printf("entry: %+v", w)
		}
	}

	kube.watchEvents()
}
