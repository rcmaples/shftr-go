package models

import (
	"time"

	"cloud.google.com/go/datastore"
)

type Appointment struct {
	Key            *datastore.Key `datastore:"__key__" json:"key"`
	Agent          string         `json:"agent,omitempty"`
	AllDay         bool           `json:"allDay,omitempty"`
	EndDate        time.Time      `json:"endDate,omitempty"`
	ExDate         string         `json:"exDate,omitempty"`
	GCalCalendarId string         `json:"gCalCalendarId,omitempty"`
	GCalEventId    string         `json:"gCalEventId,omitempty"`
	Group          string         `json:"group,omitempty"`
	Org            string         `json:"org,omitempty"`
	RRule          string         `json:"rRule,omitempty"`
	StartDate      time.Time      `json:"startDate,omitempty"`
	Title          string         `json:"title,omitempty"`
}

// type GoogleEvent struct {
// 	Agent          string    `json:"agent"`
// 	EndDate        time.Time `json:"endDate"`
// 	ExDate         string    `json:"exDate"`
// 	GCalCalendarId string    `json:"gCalCalendarId"`
// 	GCalEventId    string    `json:"gCalEventId"`
// 	Group          string    `json:"group"`
// 	Org            string    `json:"org"`
// 	RRule          string    `json:"rRule"`
// 	StartDate      time.Time `json:"startDate"`
// 	Title          string    `json:"title"`
// }
