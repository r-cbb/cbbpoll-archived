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

	srv := &http.Server{
		Handler: server.Handler(),
		Addr:    fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Serving...")

	log.Println(srv.ListenAndServe())
}