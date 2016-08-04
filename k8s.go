package main

import (
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

func watchEvents(cl kubeClient, msgr messager) error {
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

		e, ok := event.Object.(*api.Event)
		if !ok {
			continue
		}

		msgr.sendMessage(message{
			msg:       e.Message,
			obj:       e.InvolvedObject.Kind,
			name:      e.GetObjectMeta().GetName(),
			reason:    e.Reason,
			component: e.Source.Component,
		})
	}
}
