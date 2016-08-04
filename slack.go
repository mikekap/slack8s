package main

import (
	"strings"

	"github.com/nlopes/slack"
)

type messagePoster interface {
	PostMessage(string, string, slack.PostMessageParameters) (string, string, error)
}

type slackCfg struct {
	messagePoster
	channel string
	env     string
}

func (cl *slackCfg) getAttachFields(msg message) (fields []slack.AttachmentField) {
	fields = []slack.AttachmentField{
		slack.AttachmentField{
			Title: "Message",
			Value: msg.msg,
		},
		slack.AttachmentField{
			Title: "Object",
			Value: msg.obj,
			Short: true,
		},
		slack.AttachmentField{
			Title: "Name",
			Value: msg.name,
			Short: true,
		},
		slack.AttachmentField{
			Title: "Reason",
			Value: msg.reason,
			Short: true,
		},
		slack.AttachmentField{
			Title: "Component",
			Value: msg.component,
			Short: true,
		},
	}

	if cl.env != "" {
		fields = append(fields, slack.AttachmentField{
			Title: "Environment",
			Value: cl.env,
			Short: true,
		})
	}

	return fields
}

// Sends a message to the Slack channel about the Event.
func (cl *slackCfg) sendMessage(msg message) error {
	var color string
	if msg.color != "" {
		color = msg.color
	} else if strings.HasPrefix(msg.reason, "Success") {
		color = "good"
	} else if strings.HasPrefix(msg.reason, "Fail") {
		color = "danger"
	}

	params := slack.NewPostMessageParameters()

	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Color:    color,
			Fallback: msg.msg,
			Fields:   cl.getAttachFields(msg),
		},
	}

	_, _, err := cl.PostMessage(cl.channel, "", params)

	if err != nil {
		return err
	}

	return nil
}
