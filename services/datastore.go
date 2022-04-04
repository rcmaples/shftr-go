package services

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/joho/godotenv"
)

var db *datastore.Client
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

func init() {
	logger.Println("♻️  initializaing datastore...")
	db = createDataStoreClient()
}

func createDataStoreClient() *datastore.Client {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal("‼️ Error loading .env file: ", err)
	}

	gcpProjectId := os.Getenv("PROJECT_ID")

	client, err := datastore.NewClient(ctx, gcpProjectId)
	if err != nil {
		logger.Fatal("❌ error connecting to datastore: ", err)
		return nil
	}
	logger.Printf("🟢 succesfully connected to datastore client...")
	return client
}

func GetDB() *datastore.Client {
	return db
}