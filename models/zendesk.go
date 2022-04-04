package models

import (
	"time"

	"cloud.google.com/go/datastore"
)


type ZendeskConfig struct {
	Key          *datastore.Key `datastore:"__key__" json:"key"`
	Org          string         `json:"org"`
	SubDomain    string         `json:"subdomain"`
	UserString   string         `json:"userString"`
	ZendeskToken string         `json:"zendeskToken"`
}

type ZendeskGroup struct {
	Url         string    `json:"url"`
	Id          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Default     bool      `json:"default"`
	Deleted     bool      `json:"deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}