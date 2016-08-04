package main

import (
	"log"

	"k8s.io/kubernetes/pkg/api"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

type watcher interface {
	Watch(api.ListOptions) (watch.Interface, error)
}

type kubeClient interface {
	Events(string) k8s.EventInterface
}

type kubeCfg struct {
	kubeClient
	types map[watch.EventType]bool
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
			w, err = events.Watch(api.ListOptions{
				LabelSelector: labels.Everything(),
			})
			if err != nil {
				return err
			}
			continue
		}

		if cl.types != nil && !cl.types[event.Type] {
			continue
		}

		e, ok := event.Object.(*api.Event)
		if !ok {
			continue
		}

		log.Printf("received event type=%s, message=%s, reason=%s", event.Type, e.Message, e.Reason)

		msgr.sendMessage(message{
			msg:       e.Message,
			obj:       e.InvolvedObject.Kind,
			name:      e.GetObjectMeta().GetName(),
			reason:    e.Reason,
			component: e.Source.Component,
		})
	}
}
