package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shftr/helpers"
	"shftr/models"

	"cloud.google.com/go/datastore"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func saveAgents(w http.ResponseWriter, r *http.Request) {
	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
	user := helpers.UnmarshalToken(jwtContext)
	org := user.Org
	agents, err := helpers.SyncAllAgents(org)
	if err != nil {
		helpers.Logger.Println("error syncing agents: ", err)
		errorJson(w, err)
	}
	responseJson(w, http.StatusOK, agents, "agents")
}

func updateQueueShare(w http.ResponseWriter, r *http.Request) {
	var reqBody models.QueueSharePayload

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body: ", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	key, err := datastore.DecodeKey(reqBody.Key)
	if err != nil {
		helpers.Logger.Printf("error resolving agent key: %s—\n%v", reqBody.Key, err)
		errorJson(w, err, http.StatusBadRequest)
	}

	agent, err := helpers.UpdateAgentQueueShare(key, reqBody)
	if err != nil {
		helpers.Logger.Printf("error updating agent's queue share: %s—\n%v", reqBody.Key, err)
		errorJson(w, err, http.StatusInternalServerError)
	}

	responseJson(w, http.StatusOK, agent, "agent")
}

func pauseAgent(w http.ResponseWriter, r *http.Request) {
	var reqBody models.PausePayload

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body", err)
		errorJson(w, err, http.StatusBadRequest)
	}

	key, err := datastore.DecodeKey(reqBody.Key)
	if err != nil {
		helpers.Logger.Printf("error resolving agent key: %s—\n%v", reqBody.Key, err)
		errorJson(w, err, http.StatusBadRequest)
	}

	agent, err := helpers.PauseAgent(key, reqBody)
	if err != nil {
		helpers.Logger.Printf("error updating agent: %s—\n%v", reqBody.Key, err)
		errorJson(w, err, http.StatusInternalServerError)
	}

	responseJson(w, http.StatusOK, agent, "agent")
}

func findAgentByStatus(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	status := params["status"]

	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
	user := helpers.UnmarshalToken(jwtContext)
	org := user.Org

	if status == "" {
		helpers.Logger.Println("invalid status parameter")
		err := fmt.Errorf("invalid status parameter")
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	out, err := helpers.GetActiveAgents(org, status)
	if err != nil {
		helpers.Logger.Println("error getting active agents: ", err)
		errorJson(w, err)
	}

	err = responseJson(w, http.StatusOK, out, "agents")
	if err != nil {
		helpers.Logger.Println("error writing response: ", err)
		errorJson(w, err, http.StatusBadRequest)
	}
}

func updateAgents(w http.ResponseWriter, r *http.Request) {
	var reqBody []models.Agent
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	agents, err := helpers.UpdateAgents(reqBody)
	if err != nil {
		helpers.Logger.Println("error updating multiple agents", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	responseJson(w, http.StatusOK, agents, "agents")
}

func deleteAgent(w http.ResponseWriter, r *http.Request) {
	var reqBody []models.Agent
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body", err)
		errorJson(w, err, http.StatusBadRequest)
	}

	var keyStrings []string
	var bulkKeys []*datastore.Key
	for i := range reqBody {
		bulkKeys = append(bulkKeys, reqBody[i].Key)
		keyString := reqBody[i].Key.Encode()
		keyStrings = append(keyStrings, keyString)
	}
	if err := helpers.DeleteAgents(bulkKeys); err != nil {
		helpers.Logger.Println("error deleting multiple agents: ", err)
		errorJson(w, err, http.StatusInternalServerError)
	}

	helpers.Logger.Printf("deleted the following agents: %s", keyStrings)
	responseJson(w, http.StatusAccepted, keyStrings, "deleted")
}
