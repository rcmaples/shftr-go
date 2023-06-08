package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"shftr/app"
	"shftr/helpers"
	"shftr/services"
	"time"

	"github.com/joho/godotenv"
	"github.com/stephenafamo/kronika"
)

// var flgWebpack = flag.String("webpack", "http://127.0.0.1:3001", "Upstream webpack server for debug mode. If debug mode is disabled, the assets are provided from local filesystem.")
var flgWebpack = flag.String("webpack", "http://127.0.0.1:3000", "Upstream webpack server for debug mode. If debug mode is disabled, the assets are provided from local filesystem.")
var flgWebroot = flag.String("webroot", "./client/build", "Path to asset root directory for production mode. If debug mode is enabled, the assets are provided via webpack debug server.")
var flgPort = flag.Int("port", 4000, "Server port to listen on")
var flgEnv = flag.String("env", "dev", "Appliction environment (dev|prod)")

// var flgDsn = flag.String("dsn", "", "Datastore connection string")

func main() {
	flag.Parse()
	ctx := context.Background()

	services.TestCal()

	// using godotenv to load in environment variables instead of using global EXPORT
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var asset http.Handler
	if *flgEnv != "prod" {
		u, err := url.Parse(*flgWebpack)
		if err != nil {
			log.Fatal(err)
		}
		asset = app.NewDebugAssetHandler(u)
	} else {
		asset = app.NewAssetHandler(*flgWebroot)
	}

	srv := app.Router(asset)

	go func() {
		addr := fmt.Sprintf(":%d", *flgPort)
		helpers.Logger.Printf("🚀 starting server on port %d", *flgPort)
		err := http.ListenAndServe(addr, srv)
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		start, err := time.Parse(
			"2006-01-02 15:04:05",
			"2021-01-20 14:00:00",
		)
		if err != nil {
			panic(err)
		}
		interval := time.Minute * 15
		for t := range kronika.Every(ctx, start, interval) {
			helpers.SetOnline()
			helpers.Logger.Println("⏲  checking offline queue...")
			if err := helpers.EmptyOfflineQueue(); err != nil {
				helpers.Logger.Println("error emptying offline queue: ", err, t.Format("2006-01-02 15:04:00"))
			}

		}
	}()
	select {}
}
