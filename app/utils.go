package app

import (
	// "context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
)

type Secret struct {
	Secret string `json:"secret"`
	Org    string `json:"org"`
	UserId string `json:"userid"`
}

type AppStatus struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
	Version     string `json:"version"`
}

type APIHandler struct {
	Logger *log.Logger
}

// StatusHander
func statusHandler(w http.ResponseWriter, r *http.Request) {
	env := flag.Lookup("env").Value.String()

	currentStatus := AppStatus{
		Status:      "Available",
		Environment: env,
		Version:     version,
	}

	err := responseJson(w, http.StatusOK, currentStatus, "status")
	if err != nil {
		errorJson(w, err, http.StatusInternalServerError)
	}
}

// JSON Writer for errors
func errorJson(w http.ResponseWriter, err error, status ...int) {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	type jsonError struct {
		Message string `json:"message"`
		Code    int    `json:"status_code"`
	}

	e := jsonError{
		Message: err.Error(),
		Code:    statusCode,
	}

	responseJson(w, statusCode, e, "error")
}

// JSON Writer for responses
func responseJson(w http.ResponseWriter, status int, data interface{}, wrap string) error {
	wrapper := make(map[string]interface{})
	wrapper[wrap] = data
	js, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

// Some Errors
func InternalServerError(w http.ResponseWriter) {
	err := errors.New("internal server error")
	errorJson(w, err, http.StatusInternalServerError)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	err := errors.New("404 - not found")
	errorJson(w, err, http.StatusNotFound)
}
