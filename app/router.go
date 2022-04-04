package app

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const version = "0.1.0"

const org = "fullstory"

func Router(asset http.Handler) http.Handler {

	r := mux.NewRouter()

	/* PUBLIC ROUTES */
	// Status
	r.HandleFunc("/api/v1/status", statusHandler).Methods(http.MethodGet)

	// Auth
	// r.HandleFunc("/api/v1/keys/generate", generateKey).Methods(http.MethodGet)
	// r.HandleFunc("/auth/google/callback", oauthGoogleCallback).Methods(http.MethodGet)
	// r.HandleFunc("/auth/google", oauthGoogleLogin).Methods(http.MethodGet)

	// Agents
	r.HandleFunc("/api/v1/agents", listAgents).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/agents", saveAgents).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/agent/queueshare", updateQueueShare).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/agent/pause", pauseAgent).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/agent/{status}", findAgentByStatus).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/agent", deleteAgent).Methods(http.MethodDelete)
	r.HandleFunc("/api/v1/agent", updateAgents).Methods(http.MethodPatch)

	// Appointments
	r.HandleFunc("/api/v1/appointments", listAppointments).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/appointments", createAppointments).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/appointments/{apptKey}", listOneAppointment).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/appointments/{apptKey}", updateAppointment).Methods(http.MethodPatch)
	r.HandleFunc("/api/v1/appointments/{apptKey}", deleteAppointment).Methods(http.MethodDelete)

	// Tickets
	r.HandleFunc("/api/v1/history", listRecords).Methods(http.MethodGet)
	r.HandleFunc("/ap1/v1/zendesk-ticket", handleWebhook).Methods(http.MethodPost)

	// Zendesk
	r.HandleFunc("/api/v1/zendesk-config", createZendeskConfig).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/zendesk-config", getAllZendeskConfigs).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/zendesk-config/{dskey}", getZendeskConfig).Methods(http.MethodGet)

	// Not Founds
	r.PathPrefix("/api/v1/").HandlerFunc(notFound)
	r.PathPrefix("/api/v1").HandlerFunc(notFound)
	r.PathPrefix("/api/").HandlerFunc(notFound)
	r.PathPrefix("/api").HandlerFunc(notFound)

	// Serve client page
	r.PathPrefix("/").Handler(asset).Methods(http.MethodGet)

	return enableCORS(handlers.LoggingHandler(os.Stdout, r))
}