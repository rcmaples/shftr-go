package app

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const version = "0.1.0"

func Router(asset http.Handler) http.Handler {

	r := mux.NewRouter()

	/* PUBLIC ROUTES */
	// Status
	r.HandleFunc("/api/v1/status", statusHandler).Methods(http.MethodGet)

	// Auth
	r.Handle("/api/v1/keys/generate", checkToken(http.HandlerFunc(generateKey))).Methods(http.MethodGet)
	r.HandleFunc("/auth/google/callback", oauthGoogleCallback).Methods(http.MethodGet)
	r.HandleFunc("/auth/google", oauthGoogleLogin).Methods(http.MethodGet)

	/* JWT PRIVATE ROUTES ROUTES */
	// Agents
	r.Handle("/api/v1/agent/queueshare", checkToken(http.HandlerFunc(updateQueueShare))).Methods(http.MethodPatch)
	r.Handle("/api/v1/agent/pause", checkToken(http.HandlerFunc(pauseAgent))).Methods(http.MethodPatch)
	r.Handle("/api/v1/agent/{status}", checkToken(http.HandlerFunc(findAgentByStatus))).Methods(http.MethodGet)
	r.Handle("/api/v1/agent", checkToken(http.HandlerFunc(deleteAgent))).Methods(http.MethodDelete)
	r.Handle("/api/v1/agent", checkToken(http.HandlerFunc(updateAgents))).Methods(http.MethodPatch)
	r.Handle("/api/v1/agent", checkToken(http.HandlerFunc(saveAgents))).Methods(http.MethodPost)

	// Appointments
	r.Handle("/api/v1/appointments", checkToken(http.HandlerFunc(listAppointments))).Methods(http.MethodGet)
	r.Handle("/api/v1/appointments", checkToken(http.HandlerFunc(createAppointments))).Methods(http.MethodPost)
	r.Handle("/api/v1/appointments/{apptKey}", checkToken(http.HandlerFunc(listOneAppointment))).Methods(http.MethodGet)
	r.Handle("/api/v1/appointments/{apptKey}", checkToken(http.HandlerFunc(updateAppointment))).Methods(http.MethodPatch)
	r.Handle("/api/v1/appointments/{apptKey}", checkToken(http.HandlerFunc(deleteAppointment))).Methods(http.MethodDelete)

	// Tickets
	r.Handle("/api/v1/history", checkToken(http.HandlerFunc(listRecords))).Methods(http.MethodGet)
	r.Handle("/ap1/v1/zendesk-ticket", checkToken(http.HandlerFunc(handleWebhook))).Methods(http.MethodPost)

	// Zendesk
	r.Handle("/api/v1/zendesk-config", checkToken(http.HandlerFunc(createZendeskConfig))).Methods(http.MethodPost)
	r.Handle("/api/v1/zendesk-config", checkToken(http.HandlerFunc(getAllZendeskConfigs))).Methods(http.MethodGet)
	r.Handle("/api/v1/zendesk-config/{dskey}", checkToken(http.HandlerFunc(getZendeskConfig))).Methods(http.MethodGet)

	// Not Founds
	r.PathPrefix("/api/v1/").HandlerFunc(notFound)
	r.PathPrefix("/api/v1").HandlerFunc(notFound)
	r.PathPrefix("/api/").HandlerFunc(notFound)
	r.PathPrefix("/api").HandlerFunc(notFound)

	// Serve client page
	r.PathPrefix("/").Handler(asset).Methods(http.MethodGet)

	return enableCORS(handlers.LoggingHandler(os.Stdout, r))
}
