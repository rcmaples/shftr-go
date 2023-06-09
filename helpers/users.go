package helpers

import (
	"context"
	"shftr/models"
	"shftr/services"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

func GoogleHelper(gu models.GoogleUser) (models.ShftrUser, error) {
	var su models.ShftrUser

	ctx := context.Background()
	dsClient := services.GetDB()

	emailStr := gu.Email
	org := strings.Split(strings.Split(emailStr, "@")[1], ".")[0]

	// Check for duplicate entry; if found save to su.
	q := datastore.NewQuery(Users).Filter("GoogleId =", gu.Id)
	c, _ := dsClient.Count(ctx, q)
	if c > 0 {
		it := dsClient.Run(ctx, q)
		k, err := it.Next(&su)
		if err != nil {
			Logger.Println("error getting existing entry: ", err)
			return models.ShftrUser{}, err
		}
		// update su to match gu (sans createdat)
		su.GoogleId = gu.Id
		su.Email = gu.Email
		su.Name = gu.Name
		su.Picture = gu.Picture
		su.Org = org
		su.UpdatedAt = time.Now()

		ek, err := dsClient.Put(ctx, k, &su)
		if err != nil {
			Logger.Println("error updating user: ", err)
			return models.ShftrUser{}, err
		}

		// get "updatedUser" from datastore
		var updatedUser models.ShftrUser
		err = dsClient.Get(ctx, ek, &updatedUser)
		if err != nil {
			Logger.Println("error getting updated user: ", err)
			return models.ShftrUser{}, err
		}

		return updatedUser, err
	}

	// No duplicate, let's save a new user.
	su.GoogleId = gu.Id
	su.Email = gu.Email
	su.Name = gu.Name
	su.Org = org
	su.Picture = gu.Picture
	su.CreatedAt = time.Now()
	su.UpdatedAt = time.Now()

	incompleteKey := datastore.IncompleteKey("Users", nil)
	entityKey, err := dsClient.Put(ctx, incompleteKey, &su)
	if err != nil {
		Logger.Println("error creating new user: ", err)
		return models.ShftrUser{}, err
	}

	var savedUser models.ShftrUser
	err = dsClient.Get(ctx, entityKey, &savedUser)
	if err != nil {
		Logger.Println("error getting new user: ", err)
		return models.ShftrUser{}, err
	}

	return savedUser, err
}
