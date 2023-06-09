package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"shftr/helpers"
	"shftr/models"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func createAppointments(w http.ResponseWriter, r *http.Request) {
	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
	user := helpers.UnmarshalToken(jwtContext)
	org := user.Org

	var reqBody models.CreateAppointmentPayload
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	out, err := helpers.CreateAppointment(org, reqBody)
	if err != nil {
		helpers.Logger.Println("error creating appointment: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	responseJson(w, http.StatusOK, out, "appointment")
}

func listAppointments(w http.ResponseWriter, r *http.Request) {
	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
	user := helpers.UnmarshalToken(jwtContext)
	org := user.Org

	out, err := helpers.GetAllAppointments(org)
	if err != nil {
		helpers.Logger.Println("error geting all appointments: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	if err = responseJson(w, http.StatusOK, out, "appointments"); err != nil {
		helpers.Logger.Println("error writing response: ", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}
}

func listOneAppointment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apptKey := params["apptKey"]

	if apptKey == "" {
		helpers.Logger.Println("invalid appt key")
		err := fmt.Errorf("invalid appt key")
		errorJson(w, err, http.StatusBadRequest)
	}

	out, err := helpers.GetOneAppointment(apptKey)
	if err != nil {
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	if err = responseJson(w, http.StatusOK, out, "appointment"); err != nil {
		helpers.Logger.Println("error writing response: ", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}
}

func updateAppointment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apptKey := params["apptKey"]

	if apptKey == "" {
		helpers.Logger.Println("invalid appt key")
		err := fmt.Errorf("invalid appt key")
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	var reqBody models.UpdateAppointmentPayload
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		helpers.Logger.Println("error decoding request body", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}

	appt, err := helpers.UpdateAppointment(reqBody)
	if err != nil {
		helpers.Logger.Println("error updating appointment: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	if err = responseJson(w, http.StatusOK, appt, "appointment"); err != nil {
		helpers.Logger.Println("error writing response: ", err)
		errorJson(w, err, http.StatusBadRequest)
		return
	}
}

func deleteAppointment(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apptKey := params["apptKey"]

	if err := helpers.DeleteAppointment(apptKey); err != nil {
		helpers.Logger.Println("error deleting appointment: ", err)
		errorJson(w, err, http.StatusInternalServerError)
		return
	}

	responseJson(w, http.StatusAccepted, "deleted", "message")

}
