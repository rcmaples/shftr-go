package helpers

import (
	"context"
	"fmt"
	"shftr/models"
	"shftr/services"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/calendar/v3"
)

const calId = services.ShftrCal

func createGoogleEvent(evt models.Appointment) (*calendar.Event, error) {
	ctx := context.Background()
	gCalSvc := services.GetCalSvc()
	dsClient := services.GetDB()

	var agent models.Agent
	key, err := datastore.DecodeKey(evt.Agent)
	if err != nil {
		Logger.Printf("error decoding agent key: %v", err)
		return nil, err
	}
	err = dsClient.Get(ctx, key, &agent)
	if err != nil {
		Logger.Printf("error getting agent by key: %v", err)
		return nil, err
	}

	startDate := evt.StartDate.Format(time.RFC3339)
	endDate := evt.EndDate.Format(time.RFC3339)

	event := &calendar.Event{}
	if evt.RRule == "" {
		event = &calendar.Event{
			Summary: evt.Title,
			Start: &calendar.EventDateTime{
				DateTime: startDate,
				TimeZone: "Etc/UTC",
			},
			End: &calendar.EventDateTime{
				DateTime: endDate,
				TimeZone: "Etc/UTC",
			},
			Attendees: []*calendar.EventAttendee{
				{Email: agent.Email},
			},
		}
	} else {
		// the client-side scheduler component creates invalid RRULEs in certain scenarios.
		// we fix that here
		if !strings.Contains(evt.RRule, "RRULE:") {
			rrule := fmt.Sprintf("RRULE:%v", evt.RRule)
			event = &calendar.Event{
				Summary: evt.Title,
				Start: &calendar.EventDateTime{
					DateTime: startDate,
					TimeZone: "Etc/UTC",
				},
				End: &calendar.EventDateTime{
					DateTime: endDate,
					TimeZone: "Etc/UTC",
				},
				Recurrence: []string{rrule},
				Attendees: []*calendar.EventAttendee{
					{Email: agent.Email},
				},
			}
		} else {
			event = &calendar.Event{
				Summary: evt.Title,
				Start: &calendar.EventDateTime{
					DateTime: startDate,
					TimeZone: "Etc/UTC",
				},
				End: &calendar.EventDateTime{
					DateTime: endDate,
					TimeZone: "Etc/UTC",
				},
				Recurrence: []string{evt.RRule},
				Attendees: []*calendar.EventAttendee{
					{Email: agent.Email},
				},
			}
		}
	}

	out, err := gCalSvc.Events.Insert(calId, event).Do()
	if err != nil {
		Logger.Printf("err creating event in google calendar: %v", err)
		return nil, err
	}
	return out, nil
}

func ListGoogleEvents() {
	// take 'from' and 'until' date values
	// use google's calendar client to list events from shftr calendar
	// needs:
	// -- `calenderId`
	// -- `timeMin` ( from or now)
	// -- `timeMax` (until or 1 week from now)
	// -- `singleEvents` (true)
	// -- `orderBy` ("startTime")
	// send the above to google, return a slice of google events
	Logger.Println("`ListGoogleEvents` is not yet implemented...")
}

func modifyGoogleEvent(evt models.Appointment) error {
	ctx := context.Background()
	gCalSvc := services.GetCalSvc()
	dsClient := services.GetDB()

	var agent models.Agent
	key, err := datastore.DecodeKey(evt.Agent)
	if err != nil {
		Logger.Printf("error decoding agent key: %v", err)
		return err
	}
	err = dsClient.Get(ctx, key, &agent)
	if err != nil {
		Logger.Printf("error getting agent by key: %v", err)
		return err
	}

	startDate := evt.StartDate.Format(time.RFC3339)
	endDate := evt.EndDate.Format(time.RFC3339)

	var rSlice []string
	if evt.RRule != "" {
		if !strings.Contains(evt.RRule, "RRULE:") {
			rrule := fmt.Sprintf("RRULE:%s", evt.RRule)
			rSlice = append(rSlice, rrule)
		} else {
			rSlice = append(rSlice, evt.RRule)
		}
	}
	if evt.ExDate != "" {
		exDate := fmt.Sprintf("EXDATE;VALUE=DATE:%s", evt.ExDate)
		rSlice = append(rSlice, exDate)
	}

	event := &calendar.Event{
		Summary: evt.Title,
		Start: &calendar.EventDateTime{
			DateTime: startDate,
			TimeZone: "Etc/UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endDate,
			TimeZone: "Etc/UTC",
		},
		Recurrence: rSlice,
		Attendees: []*calendar.EventAttendee{
			{Email: agent.Email},
		},
	}

	_, err = gCalSvc.Events.Update(evt.GCalCalendarId, evt.GCalEventId, event).Do()
	if err != nil {
		Logger.Printf("err updating event in google calendar: %v", err)
		return err
	}
	return nil
}

func deleteGoogleEvent(evt models.Appointment) error {
	gCalSvc := services.GetCalSvc()

	Logger.Printf("Deleting event \"%s\" (%s) from %s...", evt.Title, evt.GCalEventId, evt.GCalCalendarId)

	if err := gCalSvc.Events.Delete(evt.GCalCalendarId, evt.GCalEventId).Do(); err != nil {
		return err
	}

	return nil
}
