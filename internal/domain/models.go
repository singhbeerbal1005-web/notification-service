package domain

import "time"

type Channel string

const (
	ChannelEmail Channel = "EMAIL"
	ChannelSlack Channel = "SLACK"
	ChannelInApp Channel = "IN_APP"
)

type Notification struct {
	ID             string
	Recipient      string
	Type           Channel
	TemplateName   string
	CustomTemplate string
	Payload        map[string]interface{}
	ScheduledFor   time.Time
}
