package models

import (
	"time"

	"cloud.google.com/go/datastore"
)

type AssignmentRecord struct {
	Key           *datastore.Key `datastore:"__key__" json:"key"`
	AssignedAt    time.Time      `json:"assignedAt"`
	Email         string         `json:"email"`
	GroupName     string         `json:"groupName"`
	Name          string         `json:"name"`
	Org           string         `json:"org"`
	TicketId      string         `json:"ticketId"`
	TicketUrl     string         `json:"ticketUrl"`
	ZendeskUserId int            `json:"zendeskUserId"`
}

type Ticket struct {
	Key       *datastore.Key `datastore:"__key__" json:"key"`
	GroupName string         `json:"groupName,omitempty"`
	TicketId  string         `json:"ticketId,omitempty"`
	TicketUrl string         `json:"ticketUrl,omitempty"`
	Org       string         `json:"org,omitempty"`
	DateAdded time.Time      `json:"dateAdded,omitempty"`
}