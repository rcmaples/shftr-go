package helpers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"shftr/models"
	"shftr/services"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

const zdAdminUrl = "https://%s.zendesk.com/api/v2/users?role=2"
const zdStaffUrl = "https://%s.zendesk.com/api/v2/users?permission_set=3090907"
const zdGroupsUrl = "https://%s.zendesk.com/api/v2/groups.json"

func UpdateAgents(updates []models.Agent) ([]models.Agent, error) {
	ctx := context.Background()
	dsClient := services.GetDB()

	var out []models.Agent

	for _, update := range updates {
		key := update.Key
		tx, err := dsClient.NewTransaction(ctx)
		if err != nil {
			return nil, err
		}
		var tempAgent models.Agent
		if err := tx.Get(key, &tempAgent); err != nil {
			return nil, err
		}
		tempAgent = update
		if _, err := tx.Put(key, &tempAgent); err != nil {
			return nil, err
		}

		if _, err := tx.Commit(); err != nil {
			Logger.Printf("error putting agent in transaction: %v", err)
			return nil, err
		}

		out = append(out, tempAgent)
	}
	return out, nil
}

func DeleteAgents(keys []*datastore.Key) error {
	dsClient := services.GetDB()
	ctx := context.Background()

	if len(keys) > 1 {
		if err := dsClient.DeleteMulti(ctx, keys); err != nil {
			Logger.Println("error deleting multiple agents: ", err)
			return err
		}
		return nil
	}

	key := keys[0]
	strKey := key.Encode()
	if err := dsClient.Delete(ctx, key); err != nil {
		Logger.Printf("error deleting agent: %s\n\t%v", strKey, err)
		return err
	}
	return nil
}

func SyncAllAgents(org string) ([]models.Agent, error) {

	var syncedAgents []models.Agent

	config, err := GetZendeskConfig(org)
	if err != nil {
		Logger.Println("error in SyncAllAgents: ", err)
		return nil, err
	}

	subdomain := config.SubDomain
	userString := config.UserString
	zendeskToken := config.ZendeskToken
	tokenString := fmt.Sprintf("%s:%s", userString, zendeskToken)
	b64tok := base64.StdEncoding.EncodeToString([]byte(tokenString))
	authString := fmt.Sprintf("Basic %s", b64tok)

	adminUrl := fmt.Sprintf(zdAdminUrl, subdomain)
	staffUrl := fmt.Sprintf(zdStaffUrl, subdomain)
	groupsUrl := fmt.Sprintf(zdGroupsUrl, subdomain)

	groups, err := getZendeskGroups(groupsUrl, authString)
	if err != nil {
		Logger.Println("error getting groups: ", err)
		return nil, err
	}

	admins, err := getZendeskUsers(adminUrl, authString)
	if err != nil {
		Logger.Println("error getting admins: ", err)
		return nil, err
	}

	staff, err := getZendeskUsers(staffUrl, authString)
	if err != nil {
		Logger.Println("error getting staff: ", err)
		return nil, err
	}

	adminAgents, err := saveAgents(admins, groups, org)
	if err != nil {
		Logger.Println("error saving admin agents: ", err)
		return []models.Agent{}, err
	}

	staffAgents, err := saveAgents(staff, groups, org)
	if err != nil {
		Logger.Println("error saving staff agents: ", err)
		return nil, err
	}

	syncedAgents = append(syncedAgents, adminAgents...)
	syncedAgents = append(syncedAgents, staffAgents...)

	return syncedAgents, nil
}

func GetActiveAgents(org string, status string) ([]models.Agent, error) {
	Logger.Println("🔍 retrieving active agents")

	var out []models.Agent
	ctx := context.Background()
	dsClient := services.GetDB()
	activated := false

	if status == "active" {
		activated = true
	}

	q := datastore.NewQuery(Agents).Filter("Activated =", activated)
	it := dsClient.Run(ctx, q)
	for {
		var agent models.Agent
		_, err := it.Next(&agent)
		if err == iterator.Done {
			break
		}
		if err != nil {
			Logger.Println("error fetching next agent: ", err)
			return nil, err
		}
		out = append(out, agent)
	}

	return out, nil
}

func GetAllAgents(org string) ([]models.Agent, error) {
	Logger.Println("🔍 retrieving all agents...")
	var out []models.Agent
	ctx := context.Background()
	dsClient := services.GetDB()
	q := datastore.NewQuery(Agents).Filter("Org =", org)
	it := dsClient.Run(ctx, q)
	for {
		var agent models.Agent
		_, err := it.Next(&agent)
		if err == iterator.Done {
			break
		}
		if err != nil {
			Logger.Println("error fetching next agent: ", err)
			return nil, err
		}
		out = append(out, agent)
	}

	return out, nil

}

func UpdateAgentQueueShare(key *datastore.Key, updates models.QueueSharePayload) (models.Agent, error) {
	ctx := context.Background()
	dsClient := services.GetDB()

	tx, err := dsClient.NewTransaction(ctx)
	if err != nil {
		Logger.Printf("error creating new transaction: %v", err)
		return models.Agent{}, err
	}

	var agent models.Agent
	if err := tx.Get(key, &agent); err != nil {
		Logger.Printf("error getting agent in transaction: %v", err)
		return models.Agent{}, err
	}

	agent.QueueShare.Mobile = updates.Mobile
	agent.QueueShare.SupEng = updates.SupEng
	agent.QueueShare.TechCheck = updates.TechCheck

	if _, err := tx.Put(key, &agent); err != nil {
		Logger.Printf("error putting agent in transaction: %v", err)
		return models.Agent{}, err
	}

	if _, err := tx.Commit(); err != nil {
		Logger.Printf("error committing agent in transaction: %v", err)
		return models.Agent{}, err
	}

	return agent, nil
}

func ModifyAgentStatus(key string, status bool) error {
	ctx := context.Background()
	dsClient := services.GetDB()

	dsKey, err := datastore.DecodeKey(key)
	if err != nil {
		Logger.Printf("error modifying agent %s's status to %v", key, status)
		return err
	}

	tx, err := dsClient.NewTransaction(ctx)
	if err != nil {
		Logger.Printf("error creating new transaction: %v", err)
		return err
	}

	var agent models.Agent
	if err := tx.Get(dsKey, &agent); err != nil {
		Logger.Printf("error getting agent in transaction: %v", err)
		return err
	}

	agent.Online = status

	if _, err := tx.Put(dsKey, &agent); err != nil {
		Logger.Printf("error putting agent in transaction: %v", err)
		return err
	}

	if _, err := tx.Commit(); err != nil {
		Logger.Printf("error committing agent in transaction: %v", err)
		return err
	}

	return nil
}

func PauseAgent(key *datastore.Key, updates models.PausePayload) (models.Agent, error) {
	ctx := context.Background()
	dsClient := services.GetDB()

	tx, err := dsClient.NewTransaction(ctx)
	if err != nil {
		Logger.Printf("error creating new transaction: %v", err)
		return models.Agent{}, err
	}

	var agent models.Agent
	if err := tx.Get(key, &agent); err != nil {
		Logger.Printf("error getting agent in transaction: %v", err)
		return models.Agent{}, err
	}

	agent.Paused = updates.Paused
	if _, err := tx.Put(key, &agent); err != nil {
		Logger.Printf("error putting agent in transaction: %v", err)
		return models.Agent{}, err
	}

	if _, err := tx.Commit(); err != nil {
		Logger.Printf("error committing agent in transaction: %v", err)
		return models.Agent{}, err
	}
	return agent, nil
}

func saveAgents(agents []models.ZendeskUser, groups []models.ZendeskGroup, org string) ([]models.Agent, error) {

	dsClient := services.GetDB()
	ctx := context.Background()

	var outAgents []models.Agent

	defaultQueshare := models.QueueShare{
		TechCheck: 0.0,
		SupEng:    0.0,
		Mobile:    0.0,
	}

	for i := range agents {
		var agentToSave models.Agent
		var zendeskDefaultGroupName string
		var zendeskDefaultGroupId int

		for j := range groups {
			if groups[j].Id == agents[i].DefaultGroupId {
				zendeskDefaultGroupName = groups[j].Name
				zendeskDefaultGroupId = groups[j].Id
			}
		}

		q := datastore.NewQuery(Agents).Filter("ZendeskId =", agents[i].ZendeskId)
		c, _ := dsClient.Count(ctx, q)
		if c < 1 {
			agentToSave.ZendeskId = agents[i].ZendeskId
			agentToSave.Name = agents[i].Name
			agentToSave.Email = agents[i].Email
			agentToSave.Org = org
			agentToSave.Activated = false
			agentToSave.Online = false
			agentToSave.Text = agents[i].Name
			agentToSave.QueueShare = defaultQueshare
			agentToSave.Paused = false
			agentToSave.DefaultZendeskGroupID = zendeskDefaultGroupId
			agentToSave.DefaultZendeskGroupName = zendeskDefaultGroupName
			agentToSave.CreatedAt = time.Now()

			incompleteKey := datastore.IncompleteKey(Agents, nil)
			entityKey, err := dsClient.Put(ctx, incompleteKey, &agentToSave)
			if err != nil {
				Logger.Println("error writing agent to datastore: ", err)
				return nil, err
			}

			var savedAgent models.Agent
			err = dsClient.Get(ctx, entityKey, &savedAgent)
			if err != nil {
				Logger.Println("error getting agent from datastore: ", err)
				return nil, err
			}
			outAgents = append(outAgents, savedAgent)
		}
	}

	return outAgents, nil
}

func getZendeskGroups(url string, auth string) ([]models.ZendeskGroup, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		Logger.Println("error creating group request: ", err)
		return nil, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", auth)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		Logger.Println("group response error: ", err)
		return nil, err
	}
	defer res.Body.Close()

	var body models.ZendeskGroupRes
	json.NewDecoder(res.Body).Decode(&body)

	out := body.Groups
	return out, nil
}

func getZendeskUsers(url string, auth string) ([]models.ZendeskUser, error) {
	// group ids from zendesk
	const supportEngineers = 360002978953
	const supportSpecialists = 26593617

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		Logger.Println("error creating users request: ", err)
		return []models.ZendeskUser{}, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", auth)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		Logger.Println("users response error: ", err)
		return []models.ZendeskUser{}, err
	}
	defer res.Body.Close()

	var body models.ZendeskUserRes
	json.NewDecoder(res.Body).Decode(&body)

	users := body.Users
	out := []models.ZendeskUser{}

	for i := range users {
		// only save the support specialists and engineers
		if users[i].DefaultGroupId == supportEngineers || users[i].DefaultGroupId == supportSpecialists {
			out = append(out, users[i])
		}
	}

	return out, nil
}
