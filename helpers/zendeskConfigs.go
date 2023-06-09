package helpers

import (
	"context"
	"errors"
	"shftr/models"
	"shftr/services"

	"cloud.google.com/go/datastore"
)

func SaveZendeskConfig(zd models.ZendeskConfig) (models.ZendeskConfig, error) {
	Logger.Println("🔄 saving zendesk config")

	ctx := context.Background()
	dsClient := services.GetDB()

	incompleteKey := datastore.IncompleteKey(ZendeskConfigs, nil)
	org := zd.Org

	q := datastore.NewQuery(ZendeskConfigs).Filter("Org =", org)
	c, _ := dsClient.Count(ctx, q)
	if c > 0 {
		err := errors.New("config for this org alredy exists")
		Logger.Println("❌ duplicate entry, aborting")
		return models.ZendeskConfig{}, err
	}

	entityKey, err := dsClient.Put(ctx, incompleteKey, &zd)
	if err != nil {
		Logger.Println("error puting config: ", err)
		return models.ZendeskConfig{}, err
	}

	var savedConfig models.ZendeskConfig
	err = dsClient.Get(ctx, entityKey, &savedConfig)
	if err != nil {
		Logger.Println("error getting saved config: ", err)
		return models.ZendeskConfig{}, err
	}

	return savedConfig, err
}

func GetZendeskConfig(org string) (models.ZendeskConfig, error) {
	var config models.ZendeskConfig

	Logger.Println("🔍 retrieving zendesk config")

	ctx := context.Background()
	dsClient := services.GetDB()

	q := datastore.NewQuery(ZendeskConfigs).Filter("Org =", org)
	it := dsClient.Run(ctx, q)
	_, err := it.Next(&config)
	if err != nil {
		Logger.Println("error getting zendesk config: ", err)
		return models.ZendeskConfig{}, err
	}

	return config, nil
}
