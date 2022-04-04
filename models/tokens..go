package models

import (
	"time"

	"cloud.google.com/go/datastore"
)

type APIKey struct {
	GoogleId  string    `json:"googleId"`
	UserName  string    `json:"userName"`
	Org       string    `json:"org"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
}
type Secret struct {
	Key    *datastore.Key `datastore:"__key__" json:"key"`
	Secret string         `json:"secret"`
	Org    string         `json:"org"`
	UserId string         `json:"userId"`
}

type Token struct {
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
	Id    string `json:"id"`
	Iss   string `json:"iss"`
	Name  string `json:"name"`
	Org   string `json:"org"`
}

