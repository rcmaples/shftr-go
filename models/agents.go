package models

import (
	"time"

	"cloud.google.com/go/datastore"
)


type Agent struct {
	Key                     *datastore.Key `datastore:"__key__" json:"key"`
	ZendeskId               int            `json:"zendeskId"`
	Name                    string         `json:"name"`
	Email                   string         `json:"email"`
	Org                     string         `json:"org"`
	Activated               bool           `json:"activated"`
	Online                  bool           `json:"online"`
	Color                   string         `json:"color"`
	Text                    string         `json:"text"`
	QueueShare              QueueShare     `json:"queueShare"`
	Paused                  bool           `json:"paused"`
	CreatedAt               time.Time      `json:"createdAt,omitempty"`
	DefaultZendeskGroupID   int            `json:"defaultZendeskGroupId"`
	DefaultZendeskGroupName string         `json:"defaultZendeskGroupName"`
}

type QueueShare struct {
	TechCheck float32 `json:"techcheck"`
	SupEng    float32 `json:"supeng"`
	Mobile    float32 `json:"mobile"`
}
