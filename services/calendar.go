package services

import (
	"context"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

var calSvc *calendar.Service

// const ShftrCal = "c_scvkfhjpjj8bghufa29e0r0mkg@group.calendar.google.com" // prod cal
const ShftrCal = "c_pm9slisjd0sn0ajkqt2n6g500c@group.calendar.google.com" // testing cal

func init() {
	logger.Println("♻️  initializaing calendar client...")
	calSvc = createCalService()
}

func createCalService() *calendar.Service {
	ctx := context.Background()

	ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: "shftr-go-sa@shftr-323518.iam.gserviceaccount.com",
		Scopes:          []string{calendar.CalendarEventsScope, calendar.CalendarScope},
		Subject: "rc@rcmaples.io", // Have to impersonate a real user to send invites.
	})

	if err != nil {
		logger.Fatal("❌ error impersonation credentials: ", err)
		return nil
	}

	/**
	 * Creating a new service using "Application Default Credentials"
	 * ADC is found in the GOOGLE_APPLICATION_CREDENTIALS env variable
	 * and is a path to a json file containing service account credentials
	 */
	svc, err := calendar.NewService(ctx, option.WithTokenSource(ts))

	if err != nil {
		logger.Fatal("❌ error creating calendar service: ", err)
		return nil
	}

	return svc
}

func GetCalSvc() *calendar.Service {
	return calSvc
}

func TestCal() {
	cal := GetCalSvc()

	t := time.Now().Format(time.RFC3339)
	_, err := cal.Events.List(ShftrCal).ShowDeleted(false).SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		logger.Printf("❌ error conncting to calendar: %v", err)
		return
	} else {
		logger.Printf("🟢 connected to calendar...")
	}
}
