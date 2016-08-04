package main

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type whitelistEntry struct {
	EventType string `json:"eventType,omitempty"`
	Msg       string `json:"msg,omitempty"`
	Obj       string `json:"obj,omitempty"`
	Name      string `json:"name,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Component string `json:"component,omitempty"`
}

type whitelist []whitelistEntry

func (wl whitelist) accepts(msg message) bool {
	if wl == nil {
		return true
	}

	for _, entry := range wl {
		if entry.EventType != "" && entry.EventType != msg.eventType {
			continue
		}

		if entry.Msg != "" && entry.Msg != msg.msg {
			continue
		}

		if entry.Obj != "" && entry.Obj != msg.obj {
			continue
		}

		if entry.Name != "" && entry.Name != msg.name {
			continue
		}

		if entry.Reason != "" && entry.Reason != msg.reason {
			continue
		}

		if entry.Component != "" && entry.Component != msg.component {
			continue
		}

		return true
	}

	return false
}

type kubeClient interface {
	Events(string) k8s.EventInterface
}

type kubeCfg struct {
	kubeClient
	types     map[watch.EventType]bool
	whitelist whitelist
}

func (cl *kubeCfg) watchEvents(msgr messager) error {
	events := cl.Events(api.NamespaceAll)

	w, err := events.Watch(api.ListOptions{
		LabelSelector: labels.Everything(),
	})
	if err != nil {
		return err
	}

	for {
		event, ok := <-w.ResultChan()
		if !ok {
			log.Printf("event channel closed, try reconnecting")
			w, err = events.Watch(api.ListOptions{
				LabelSelector: labels.Everything(),
			})
			if err != nil {
				return err
			}
			continue
		}

		send := true
		if cl.types != nil && !cl.types[event.Type] {
			send = false
		}

		e, ok := event.Object.(*api.Event)
		if !ok {
			continue
		}

		msg := message{
			msg:       e.Message,
			obj:       e.InvolvedObject.Kind,
			name:      e.GetObjectMeta().GetName(),
			reason:    e.Reason,
			component: e.Source.Component,
			count:     int(e.Count),
			eventType: string(event.Type),
		}

		if !cl.whitelist.accepts(msg) {
			send = false
		}

		log.Printf(
			"event type=%s, message=%s, reason=%s, send=%v",
			event.Type,
			e.Message,
			e.Reason,
			send,
		)

		if send {
			msgr.sendMessage(msg)
		}
	}
}
