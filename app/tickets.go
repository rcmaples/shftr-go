package app

import (
	"encoding/json"
	"net/http"
	"shftr/helpers"
	"shftr/models"
)

func listRecords(w http.ResponseWriter, r *http.Request) {
	out, err := helpers.GetHistoricalRecords(org)
	if err != nil {
		helpers.Logger.Printf("error getting historical records: %v", err)
		errorJson(w, err, http.StatusInternalServerError)
	}
	responseJson(w, http.StatusOK, out, "tickets")
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var reqBody models.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		helpers.Logger.Printf("error decoding request body: %v", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	if err := helpers.TicketHandler(reqBody, org); err != nil {
		helpers.Logger.Printf("error assigning the ticket: %v", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}
	responseJson(w, http.StatusOK, reqBody, "ticket")
}
