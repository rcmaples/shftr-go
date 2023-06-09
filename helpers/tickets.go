package helpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"reflect"
	"shftr/models"
	"shftr/services"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func GetHistoricalRecords(org string) ([]models.AssignmentRecord, error) {
	var out []models.AssignmentRecord
	ctx := context.Background()
	dsClient := services.GetDB()
	when := time.Now().AddDate(0, -1, 0).UTC()
	q := datastore.NewQuery(AssignedTickets).
		Filter("Org =", org).
		Filter("AssignedAt >", when)

	it := dsClient.Run(ctx, q)
	for {
		var record models.AssignmentRecord
		_, err := it.Next(&record)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}

func TicketHandler(ticket models.WebhookPayload, org string) error {
	Logger.Printf("handling ticket %s...", ticket.TicketId)
	cfg, err := GetZendeskConfig(org)
	if err != nil {
		Logger.Printf("error getting zendesk config: %v", err)
		return err
	}

	availableAgents, err := findAvailableAgents(org)
	if err != nil {
		Logger.Printf("error getting available agents: %v", err)
		return err
	}
	Logger.Printf("%d agent(s) available:", len(availableAgents))
	for i := range availableAgents {
		Logger.Printf("\t%s (%d)", availableAgents[i].Name, availableAgents[i].ZendeskId)
	}

	if len(availableAgents) < 1 {
		Logger.Printf("sending ticket %s to be added to offline queue...", ticket.TicketId)
		if err := addToOfflineQueue(ticket, cfg.Org); err != nil {
			Logger.Printf("error adding ticket %s to offline queue: %v", ticket.TicketId, err)
			return err
		}
		return nil
	}

	var group string
	switch ticket.GroupName {
	case "Support Engineers":
		group = "SupEng"
	case "Tech Check":
		group = "TechCheck"
	case "Native Mobile":
		group = "Mobile"
	default:
		Logger.Println("unknown ticket group")
		err = errors.New("unknown ticket group")
		return err
	}

	assignee, avail := randomizer(availableAgents, group)
	Logger.Printf("availability: %d -- assignee: %s (%d)", avail, assignee.Name, assignee.ZendeskId)

	if avail < 1 {
		Logger.Printf("no availability - adding ticket (%s) to offline queue...", ticket.TicketId)
		if err := addToOfflineQueue(ticket, org); err != nil {
			Logger.Printf("error adding ticket %s to offline queue: %v", ticket.TicketId, err)
			return err
		}
		return nil
	}

	Logger.Printf("attempting to assign ticket %s to %s (%d)...", ticket.TicketId, assignee.Name, assignee.ZendeskId)
	if err = assignTicket(assignee, ticket, cfg); err != nil {
		Logger.Printf("error assigning ticket %s to agent %s (%d):\n\t%v", ticket.TicketId, assignee.Name, assignee.ZendeskId, err)
		return err
	}

	Logger.Println("finished handling ticket")
	return nil
}

func findAvailableAgents(org string) ([]models.Agent, error) {
	Logger.Println("checking for available agents...")
	ctx := context.Background()
	dsClient := services.GetDB()
	var out []models.Agent

	q := datastore.NewQuery(Agents).
		Filter("Org =", org).
		Filter("Online =", true).
		Filter("Activated =", true).
		Filter("Paused =", false).
		Filter("DefaultZendeskGroupName =", "Support Engineers")

	it := dsClient.Run(ctx, q)
	for {
		var agent models.Agent
		_, err := it.Next(&agent)
		if err == iterator.Done {
			break
		}
		if err != nil {
			Logger.Printf("error fetching next agent: %v", err)
			return nil, err
		}
		out = append(out, agent)
	}

	return out, nil
}

func assignTicket(agent models.Agent, ticket models.WebhookPayload, cfg models.ZendeskConfig) error {
	Logger.Printf("beginning to assign ticket %s to %s (%d)...", ticket.TicketId, agent.Name, agent.ZendeskId)

	payloadStr := fmt.Sprintf(`{"ticket":{"assignee_email": "%s"}}`, agent.Email)
	payload := strings.NewReader(payloadStr)

	sub := cfg.SubDomain
	user := cfg.UserString
	tok := cfg.ZendeskToken
	tokStr := fmt.Sprintf("%s:%s", user, tok)
	b64tok := base64.StdEncoding.EncodeToString([]byte(tokStr))
	authStr := fmt.Sprintf("Basic %s", b64tok)

	tickId := ticket.TicketId
	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/tickets/%s.json", sub, tickId)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, url, payload)
	if err != nil {
		Logger.Println("error creating PUT ticket request: ", err)
		return err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authStr)
	Logger.Printf("sending ticket %s to zendesk for assignment...", ticket.TicketId)
	res, err := client.Do(req)
	if err != nil {
		Logger.Println("error assigning ticket: ", err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(string(body))

	Logger.Println("zendesk api response: ", res.Status)
	Logger.Printf("succesfully assigned ticket %s to %s (%d)", tickId, agent.Name, agent.ZendeskId)

	Logger.Println("logging ticket in datastore...")
	record := models.AssignmentRecord{
		AssignedAt:    time.Now(),
		Email:         agent.Email,
		GroupName:     ticket.GroupName,
		Name:          agent.Name,
		Org:           agent.Org,
		TicketId:      ticket.TicketId,
		TicketUrl:     ticket.TicketUrl,
		ZendeskUserId: agent.ZendeskId,
	}

	Logger.Println("handing off to `logTicket`...")
	if err := logTicket(record); err != nil {
		Logger.Printf("error logging ticket %s: %v", ticket.TicketId, err)
		return err
	}

	return nil
}

func logTicket(rec models.AssignmentRecord) error {
	Logger.Printf("logging ticket %s to datastore...", rec.TicketId)
	ctx := context.Background()
	dsClient := services.GetDB()
	ik := datastore.IncompleteKey(AssignedTickets, nil)
	_, err := dsClient.Put(ctx, ik, &rec)
	if err != nil {
		Logger.Println("error writing record to datastore: ", err)
		return err
	}
	Logger.Println("ticket logged.")
	return nil
}

func addToOfflineQueue(payload models.WebhookPayload, org string) error {
	Logger.Printf("attempting to add ticket %s to the offline queue...", payload.TicketId)
	var ticket models.Ticket

	jp, err := json.Marshal(payload)
	if err != nil {
		Logger.Println("unable to marshal json:", err)
		return err
	}

	err = json.Unmarshal(jp, &ticket)
	if err != nil {
		Logger.Println("unable to unmarshal json: ", err)
		return err
	}

	ticket.Org = org
	ticket.DateAdded = time.Now()

	ctx := context.Background()
	dsClient := services.GetDB()

	q := datastore.NewQuery(OfflineTickets).Filter("TicketId =", ticket.TicketId)
	c, _ := dsClient.Count(ctx, q)
	if c > 0 {
		Logger.Printf("ticket %s is already in the offline queue, aborting...", ticket.TicketId)
		return nil
	}

	ik := datastore.IncompleteKey(OfflineTickets, nil)

	_, err = dsClient.Put(ctx, ik, &ticket)
	if err != nil {
		Logger.Printf("error saving ticket %s to offline queue: %v", ticket.TicketId, err)
		return err
	}

	Logger.Printf("ticket %s added to the offline queue...", ticket.TicketId)
	return nil
}

func EmptyOfflineQueue() error {
	var tickets []models.Ticket
	ctx := context.Background()
	dsClient := services.GetDB()
	q := datastore.NewQuery(OfflineTickets)

	_, err := dsClient.GetAll(ctx, q, &tickets)
	if err != nil {
		Logger.Println("error getting offline tickets from datastore: ", err)
		return err
	}

	Logger.Printf("%d tickets in the offline queue...", len(tickets))

	if len(tickets) < 1 {
		Logger.Println("no tickets in the offline queue. aborting.")
		return nil
	}

	for _, ticket := range tickets {
		org := ticket.Org
		var payload = models.WebhookPayload{
			TicketId:  ticket.TicketId,
			GroupName: ticket.GroupName,
		}

		Logger.Printf("sending offline ticket %s to `ticketHandler`...", ticket.TicketId)
		if err := TicketHandler(payload, org); err != nil {
			Logger.Printf("error assigning ticket %s from the offline queue: %v", ticket.TicketId, err)
			return err
		}

		Logger.Printf("removing ticket %s from the offline queue...", ticket.TicketId)
		if err := dsClient.Delete(ctx, ticket.Key); err != nil {
			Logger.Printf("error removing ticket %s (%v) from the offline queue: %v", ticket.TicketId, ticket.Key, err)
			return err
		}
	}
	return nil
}

func randomizer(agents []models.Agent, group string) (models.Agent, int) {
	Logger.Println("randomly choosing an available agent...")
	var totalAvailability float64
	var seed float64
	var top float64
	var bottom float64

	seed = 0.00
	top = 0.00
	bottom = 0.00
	totalAvailability = 0.00

	for _, agent := range agents {
		groups := agent.QueueShare
		r := reflect.ValueOf(&groups)
		v := reflect.Indirect(r).FieldByName(group)
		s := float64(v.Float())
		totalAvailability += s
	}

	if totalAvailability == 0 {
		return models.Agent{}, 0
	}

	if totalAvailability > 0 {
		seed = math.Floor(rand.Float64() * totalAvailability)
	}
	Logger.Printf("total availability for group %s: %f", group, totalAvailability)

	for _, agent := range agents {
		groups := agent.QueueShare
		r := reflect.ValueOf(&groups)
		v := reflect.Indirect(r).FieldByName(group)
		s := float64(v.Float())

		top = bottom + s
		if seed >= bottom && seed < top {
			Logger.Printf("%s (%d) has been randomly chosen...", agent.Name, agent.ZendeskId)
			return agent, 1
		}
		bottom = top
	}

	Logger.Println("no one was chosen...")
	return models.Agent{}, 0
}
