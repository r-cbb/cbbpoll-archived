package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/r-cbb/cbbpoll/internal/app"
	"github.com/r-cbb/cbbpoll/internal/db"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Initializing server...")

	server := app.NewServer()
	var err error
	server.Db, err = db.NewDatastoreClient("cbbpoll")

	log.Println("\tDatastoreClient initialized")
	if err != nil {
		log.Fatal(err.Error())
		panic(err.Error())
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("\tDefaulting to port %s", port)
	} else {
		log.Printf("\tUsing port %s from environment variable", port)
	}

	server.TokenAuth, err = app.InitJwtAuth("jwtRS256.key", "jwtRS256.key.pub")
	if err != nil {
		log.Printf("error initializing JWT authentication: %s", err.Error())
	} else {
		server.AuthRoutes()
	}

	// TODO: flag to enable TLS

	srv := &http.Server{
		Handler: server,
		Addr:    fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Serving...")

	log.Println(srv.ListenAndServe())
}
