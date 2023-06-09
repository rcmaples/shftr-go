package helpers

import (
	"context"
	"fmt"
	"shftr/models"
	"shftr/services"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	rrule "github.com/teambition/rrule-go"
	"google.golang.org/api/iterator"
)

func SetOnline() {
	Logger.Println("📇 checking online status...")

	activeAgentKeyStrs, err := activeAgentKeys()
	if err != nil {
		Logger.Fatalf("❌ error getting active agents: %v", err)
	}

	onlineIds := isOnline()
	Logger.Printf("🔋 setting the following ids online: %v", onlineIds)

	for _, key := range activeAgentKeyStrs {
		if indexOf(key, onlineIds) != -1 {
			if err := ModifyAgentStatus(key, true); err != nil {
				Logger.Printf("🙊 error setting agent %s to online:\n%v", key, err)
			}
		} else {
			Logger.Printf("🪫 setting the following id ofline: %v", key)
			if err := ModifyAgentStatus(key, false); err != nil {
				Logger.Printf("🙊 error setting agent %s to online:\n%v", key, err)
			}
		}
	}
}

func isOnline() []string {
	now := time.Now()
	activeAgentKeyStrs, err := activeAgentKeys()
	if err != nil {
		Logger.Fatalf("❌ error finding active agents: %v", err)
	}
	activeAppts, err := findAgentAppointments(activeAgentKeyStrs)
	if err != nil {
		Logger.Fatalf("❌ error finding active agents: %v", err)
	}
	var onlineIds []string

	for _, appt := range activeAppts {
		duration := appt.EndDate.Sub(appt.StartDate)

		var apptRule *rrule.RRule
		var latestShiftStart time.Time
		var shiftEnd time.Time
		excluded := false

		startStr := appt.StartDate.Format(time.RFC3339)
		startStr = strings.Replace(strings.Replace(startStr, "-", "", -1), ":", "", -1)

		// Populate excluded dates
		var exDates []time.Time
		if appt.ExDate != "" {
			exDateSlice := strings.Split(appt.ExDate, ",")
			for _, d := range exDateSlice {
				yyyy := substr(d, 0, 4)
				mm := substr(d, 4, 2)
				dd := substr(d, 6, 2)
				// t := substr(d, 8, 1)
				hh := substr(d, 9, 2)
				mi := substr(d, 11, 2)
				ss := substr(d, 13, 2)
				ts := fmt.Sprintf("%s-%s-%sT%s:%s:%sZ", yyyy, mm, dd, hh, mi, ss)
				et, _ := time.Parse(time.RFC3339, ts)
				exDates = append(exDates, et)
			}
		}
		if len(exDates) > 0 {
			for _, ed := range exDates {
				edShiftStart := ed
				edShiftEnd := ed.Add(duration)
				if edShiftStart.Before(now) && edShiftEnd.After(now) {
					excluded = true
				}
			}
		}
		if appt.RRule == "" || appt.StartDate.After(now) {
			latestShiftStart = appt.StartDate
			shiftEnd = appt.EndDate
		} else {
			if strings.Contains(appt.RRule, "RRULE:") {
				ruleString := fmt.Sprintf("DTSTART:%s\n%s", startStr, appt.RRule)
				apptRule, _ = rrule.StrToRRule(ruleString)
			} else if !strings.Contains(appt.RRule, "RRULE:") {
				ruleString := fmt.Sprintf("DSTART:%s\nRRULE:%s", startStr, apptRule)
				apptRule, _ = rrule.StrToRRule(ruleString)
			}
			latestShiftStart = apptRule.Before(now, true)
			shiftEnd = latestShiftStart.Add(duration)
		}
		if latestShiftStart.Before(now) && now.Before(shiftEnd) && !excluded {
			onlineIds = append(onlineIds, appt.Agent)
		}
	}
	return onlineIds
}

func findAgentAppointments(keyStrs []string) ([]models.Appointment, error) {
	ctx := context.Background()
	dsClient := services.GetDB()

	var out []models.Appointment

	for _, ks := range keyStrs {
		q := datastore.NewQuery(Appointments).Filter("Agent =", ks)
		it := dsClient.Run(ctx, q)
		for {
			var appt models.Appointment
			_, err := it.Next(&appt)
			if err == iterator.Done {
				break
			}
			if err != nil {
				Logger.Println("error fetching next appointment: ", err)
				return nil, err
			}
			out = append(out, appt)
		}
	}
	return out, nil
}

func activeAgentKeys() ([]string, error) {
	ctx := context.Background()
	dsClient := services.GetDB()

	var agentKeyStrs []string

	q := datastore.NewQuery(Agents).Filter("Activated =", true)
	it := dsClient.Run(ctx, q)
	for {
		var agent models.Agent
		_, err := it.Next(&agent)
		if err == iterator.Done {
			break
		}
		if err != nil {
			Logger.Printf("error fetching active agent: %v", err)
			return nil, err
		}
		agentKeyStrs = append(agentKeyStrs, agent.Key.Encode())
	}

	return agentKeyStrs, nil
}
