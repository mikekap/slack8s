package main

import (
	"log"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/cache"
	"k8s.io/kubernetes/pkg/controller/framework"
	k8s "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/kubernetes/pkg/util/wait"
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

type kubeCfg struct {
	*k8s.Client
	whitelist whitelist
	msgr messager
}

func (cl *kubeCfg) onUpdate(eventType watch.EventType, obj interface{}) {
	e, ok := obj.(*api.Event)
	if !ok {
		return
	}

	msg := message{
		msg:       e.Message,
		obj:       e.InvolvedObject.Kind,
		name:      e.GetObjectMeta().GetName(),
		reason:    e.Reason,
		component: e.Source.Component,
		count:     int(e.Count),
		eventType: string(eventType),
	}

	tooOld := time.Now().Sub(e.LastTimestamp.Time) > 10 * time.Minute

	send := cl.whitelist.accepts(msg) && !tooOld
	log.Printf(
		"event type=%s, message=%s, reason=%s, send=%v",
		eventType,
		e.Message,
		e.Reason,
		send,
	)

	if send {
		cl.msgr.sendMessage(msg)
	}
}

func (cl *kubeCfg) watchEvents() {
	watchlist := cache.NewListWatchFromClient(cl, "events", api.NamespaceAll, fields.Everything())
	resyncPeriod := 30 * time.Minute

	funcs := framework.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cl.onUpdate(watch.Added, obj);
		},
		UpdateFunc: func(before, after interface{}) {
			//cl.onUpdate(watch.Modified, after);
		},
	}

	_, eController := framework.NewInformer(
		watchlist,
		&api.Event{},
		resyncPeriod,
		funcs,
	)
	eController.Run(wait.NeverStop)
}
