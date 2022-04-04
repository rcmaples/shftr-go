package models

import "time"

type UpdateAppointmentPayload struct {
	Key            string    `json:"key"`
	Agent          string    `json:"agent"`
	AllDay         bool      `json:"allDay,omitempty"`
	GCalCalendarId string    `json:"gCalCalendarId,omitempty"`
	GCalEventId    string    `json:"gCalEventId,omitempty"`
	Group          string    `json:"group,omitempty"`
	EndDate        time.Time `json:"endDate,omitempty"`
	ExDate         string    `json:"exDate,omitempty"`
	Org            string    `json:"org,omitempty"`
	RRule          string    `json:"rRule,omitempty"`
	StartDate      time.Time `json:"startDate,omitempty"`
	Title          string    `json:"title,omitempty"`
}

type CreateAppointmentPayload struct {
	Agent          string    `json:"agent"`
	AllDay         bool      `json:"allDay,omitempty"`
	EndDate        time.Time `json:"endDate"`
	ExDate         string    `json:"exDate,omitempty"`
	RRule          string    `json:"rRule,omitempty"`
	StartDate      time.Time `json:"startDate"`
	Title          string    `json:"title"`
}
type PausePayload struct {
	Key    string `json:"key"`
	Paused bool   `json:"paused"`
}

type QueueSharePayload struct {
	Key       string  `json:"key"`
	Name      string  `json:"name"`
	TechCheck float32 `json:"techcheck"`
	SupEng    float32 `json:"supeng"`
	Mobile    float32 `json:"mobile"`
	EditMode  bool    `json:"isEditMode,omitempty"`
}

type WebhookPayload struct {
	TicketId      string `json:"ticketId"`
	AssigneeEmail string `json:"assigneeEmail,omitempty"`
	GroupName     string `json:"groupName"`
	TicketStatus  string `json:"ticketStatus"`
	TicketUrl     string `json:"ticketUrl"`
}

type ZendeskGroupRes struct {
	Count        int            `json:"count"`
	NextPage     string         `json:"next_page"`
	PreviousPage string         `json:"previous_page"`
	Groups       []ZendeskGroup `json:"groups"`
}

type ZendeskUserRes struct {
	Count        int           `json:"count"`
	NextPage     string        `json:"next_page"`
	PreviousPage string        `json:"previous_page"`
	Users        []ZendeskUser `json:"users"`
}
