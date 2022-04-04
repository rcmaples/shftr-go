package helpers

import (
	"context"
	"encoding/json"
	"shftr/models"
	"shftr/services"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func CreateAppointment(org string, payload models.CreateAppointmentPayload) (models.Appointment, error) {
	Logger.Println("✍️  creating appointment...")

	var appt models.Appointment
	var out models.Appointment
	var agent models.Agent
	ctx := context.Background()
	dsClient := services.GetDB()

	keyStr := payload.Agent
	dsKey, err := datastore.DecodeKey(keyStr)
	if err != nil {
		Logger.Printf("error decoding key %s: %v", keyStr, err)
		return models.Appointment{}, err
	}

	if err = dsClient.Get(ctx, dsKey, &agent); err != nil {
		Logger.Println("error fetching agent: ", err)
		return models.Appointment{}, err
	}

	jp, err := json.Marshal(payload)
	if err != nil {
		Logger.Println("unable to marshal json: ", err)
		return models.Appointment{}, err
	}

	err = json.Unmarshal(jp, &appt)
	if err != nil {
		Logger.Println("unable to unmarshal json: ", err)
		return models.Appointment{}, err
	}

	appt.Group = agent.DefaultZendeskGroupName
	appt.Org = org

	// Create the event in Google Calendar
	gEvt, err := createGoogleEvent(appt)
	if err != nil {
		Logger.Println("error creating google event: ", err)
		return models.Appointment{}, err
	}

	// assign gcal ids to the appointment before saving to datastore
	appt.GCalCalendarId = services.ShftrCal
	appt.GCalEventId = gEvt.Id

	// save appointment in datastore
	ik := datastore.IncompleteKey(Appointments, nil)
	ek, err := dsClient.Put(ctx, ik, &appt)
	if err != nil {
		Logger.Println("error putting appointment: ", err)
		return models.Appointment{}, err
	}

	// get the appointment from datastore and store it in `out`
	err = dsClient.Get(ctx, ek, &out)
	if err != nil {
		Logger.Println("error retrieving save appointment: ", err)
		return models.Appointment{}, err
	}

	return out, nil
}

func GetAllAppointments(org string) ([]models.Appointment, error) {
	Logger.Println("🔍 📅 retrieving apointments...")

	var out []models.Appointment
	ctx := context.Background()
	dsClient := services.GetDB()

	q := datastore.NewQuery(Appointments).Filter("Org =", org)
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

	return out, nil
}

func GetOneAppointment(key string) (models.Appointment, error) {
	ctx := context.Background()
	dsClient := services.GetDB()
	var appt models.Appointment

	dsKey, err := datastore.DecodeKey(key)
	if err != nil {
		Logger.Printf("error decoding key %s: %v", key, err)
		return models.Appointment{}, err
	}

	if err = dsClient.Get(ctx, dsKey, &appt); err != nil {
		Logger.Println("error fetching appointment: ", err)
		return models.Appointment{}, err
	}

	return appt, nil
}

func UpdateAppointment(payload models.UpdateAppointmentPayload) (models.Appointment, error) {
	Logger.Println("✍️  updating appointment...")

	ctx := context.Background()
	dsClient := services.GetDB()

	key, err := datastore.DecodeKey(payload.Key)
	if err != nil {
		Logger.Printf("error decoding key %s: %v", key, err)
		return models.Appointment{}, err
	}

	tx, err := dsClient.NewTransaction(ctx)
	if err != nil {
		Logger.Printf("error creating new transaction: %v", err)
		return models.Appointment{}, err
	}

	var tempAppt models.Appointment
	if err := tx.Get(key, &tempAppt); err != nil {
		Logger.Printf("error getting appointment in transaction: %v", err)
		return models.Appointment{}, err
	}

	jp, err := json.Marshal(payload)
	if err != nil {
		Logger.Println("unable to marshal json: ", err)
		return models.Appointment{}, err
	}

	err = json.Unmarshal(jp, &tempAppt)
	if err != nil {
		Logger.Println("unable to unmarshal json: ", err)
		return models.Appointment{}, err
	}

	if err = modifyGoogleEvent(tempAppt); err != nil {
		_ = tx.Rollback()
		return models.Appointment{}, err
	}

	if _, err := tx.Put(key, &tempAppt); err != nil {
		Logger.Printf("error putting appointment in transaction: %v", err)
		_ = tx.Rollback()
		return models.Appointment{}, err
	}

	if _, err := tx.Commit(); err != nil {
		Logger.Printf("error committing appointment in transaction: %v", err)
		_ = tx.Rollback()
		return models.Appointment{}, err
	}

	return tempAppt, nil
}

func DeleteAppointment(key string) error {
	Logger.Println("🗑  deleting appointment...")
	ctx := context.Background()
	dsClient := services.GetDB()

	var evt models.Appointment

	dskey, err := datastore.DecodeKey(key)
	if err != nil {
		Logger.Printf("error decoding key %s: %v", key, err)
		return err
	}

	if err = dsClient.Get(ctx, dskey, &evt); err != nil {
		Logger.Printf("error retrieving appointment to delete: %v", err)
		return err
	}

	if err = deleteGoogleEvent(evt); err != nil {
		Logger.Printf("error deleting appointment from google: %v", err)
		return err
	}

	if err = dsClient.Delete(ctx, dskey); err != nil {
		Logger.Printf("error deleting appointment from datastore: %v", err)
		return err
	}

	return nil
}
