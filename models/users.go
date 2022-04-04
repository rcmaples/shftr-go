package models

import "time"

type GoogleUser struct {
	Id      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type ShftrUser struct {
	GoogleId  string    `json:"googleId"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Org       string    `json:"org"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"date,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type ZendeskUser struct {
	ZendeskId      int    `json:"id"`
	DefaultGroupId int    `json:"default_group_id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
}