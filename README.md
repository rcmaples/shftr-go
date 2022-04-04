<p align="center">
  <img src="./go-shftr.png" alt="shftr go logo">
</p>

# shftr in go

React front end is in `./client`

Go backend is in `./server`

Getting Started

- Launch datastore:

  from inside `/shftr`:

    - `gcloud beta emulators datastore start --data-dir=datastore --project=go-shftr`

- Launch the backend:

  from inside `/shftr`:

    - `go run main.go --env prod|dev`
    - (dev looks for dev build of front end, prod looks for production build)

- Launch the frontend:

  from inside `/shftr/client`:

    - `npm run start` -- build the dev version with hotreloading; port 3000 by default (conflicts with datastore gui)

    - `npm run build` -- builds the front end for the shftr backend to serve staticly

- datastore emulator gui:

  from **outside** of `/shftr`:

    - `google-cloud-gui --port 3000` (requires google-cloud-gui to be installed via `npm i -g`)


---

### Go dependencies:

- [godotevnt](https://github.com/joho/godotenv) (for managing environment variables)
- [Google Cloud Datastore](https://pkg.go.dev/cloud.google.com/go/datastore) (database)
- [golang-jwt v4](https://github.com/golang-jwt/jwt/v4) (jwt creation and validation)
- [gorilla handlers](https://github.com/gorilla/handlers) (http handlers from gorilla)
- [gorilla mux](https://github.com/gorilla/mux) (gorilla mux for http routing)
- [kronika](https://github.com/stephenafamo/kronika) (cron scheduling)

---

### React dependencies:

- more to come.

