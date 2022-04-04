package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"shftr/helpers"
	"shftr/models"
	"shftr/services"

	"cloud.google.com/go/datastore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func createZendeskConfig(w http.ResponseWriter, r *http.Request) {
	var i models.ZendeskConfig

	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
	user := helpers.UnmarshalToken(jwtContext)

	err := json.NewDecoder(r.Body).Decode(&i)
	if err != nil {
		helpers.Logger.Println("error decoding json into zendesk config: ", err)
		err = errors.New(err.Error() + errors.New(": invalid json").Error())
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	var contentError error
	if i.SubDomain == "" {
		contentError = errors.New("missing subdomain")
	}
	if i.UserString == "" {
		contentError = errors.New("missing userString")
	}
	if i.ZendeskToken == "" {
		contentError = errors.New("missing zendeskToken")
	}
	if contentError != nil {
		errorJson(w, contentError, http.StatusBadRequest)
		return
	}

	i.Org = user.Org
	returnConfig, err := helpers.SaveZendeskConfig(i)

	if err != nil {
		if err.Error() == "config for this org alredy exists" {
			errorJson(w, err, http.StatusBadRequest)
			return
		}
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	// sanitize returnConfig so we don't send the token back to the client.
	var out struct {
		Org        string         `json:"org"`
		Subdomain  string         `json:"subdomain"`
		UserString string         `json:"userString"`
		Key        *datastore.Key `datastore:"__key__" json:"key"`
	}

	out.Org = returnConfig.Org
	out.Subdomain = returnConfig.SubDomain
	out.UserString = returnConfig.UserString
	out.Key = returnConfig.Key

	responseJson(w, http.StatusCreated, out, "config")
}

func getAllZendeskConfigs(w http.ResponseWriter, r *http.Request) {
	var configs []models.ZendeskConfig

	ctx := context.Background()

	dsClient := services.GetDB()

	q := datastore.NewQuery(helpers.ZendeskConfigs)
	keys, err := dsClient.GetAll(ctx, q, &configs)
	if err != nil {
		helpers.Logger.Println("error getting all configs: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	err = responseJson(w, http.StatusOK, keys, "configKeys")
	if err != nil {
		helpers.Logger.Println("error sending response json: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}
}

func getZendeskConfig(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	dskey := params["dskey"]

	if dskey == "" {
		helpers.Logger.Println("invalid id parameter")
		err := fmt.Errorf("something went wrong")
		errorJson(w, err)
		return
	}

	out, err := helpers.GetZendeskConfig(dskey)
	if err != nil {
		helpers.Logger.Println("error getting zendesk config: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	err = responseJson(w, http.StatusOK, out, "zendeskConfig")
	if err != nil {
		helpers.Logger.Println("error writing response json: ", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

}
