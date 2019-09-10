package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"

	_ "github.com/r-cbb/cbbpoll/docs"
	"github.com/r-cbb/cbbpoll/internal/app"
	"github.com/r-cbb/cbbpoll/internal/auth"
	"github.com/r-cbb/cbbpoll/internal/db"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Initializing server...")

	server := app.NewServer()
	var err error

	// Setup Datastore connection
	server.Db, err = db.NewDatastoreClient("cbbpoll")

	log.Println("\tDatastoreClient initialized")
	if err != nil {
		log.Fatal(err.Error())
		panic(err.Error())
	}

	// Setup JWT Auth
	setupAuth(server)

	// Setup HTTP Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("\tDefaulting to port %s", port)
	} else {
		log.Printf("\tUsing port %s from environment variable", port)
	}

	// Setup reddit client
	// TODO read from config
	server.RedditClient = app.NewRedditClient("https://oauth.reddit.com/api/v1")

	// Enable CORS for Swagger-UI
	// TODO behind config flag as well?
	c := cors.New(cors.Options{
		Debug: false,
		AllowedHeaders:[]string{"*"},
		AllowedOrigins:[]string{"*"},
		AllowedMethods:[]string{},
		MaxAge:1000,
	})

	handler := c.Handler(server)

	// TODO: flag to enable TLS
	srv := &http.Server{
		Handler: handler,
		Addr:    fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// TODO: config
	server.SetHost("http://localhost:8000")

	log.Println("Serving...")
	log.Println(srv.ListenAndServe())
}

func setupAuth(server *app.Server) {
	keyFile, err := os.Open("jwtRS256.key")
	if err != nil {
		log.Fatalf("error opening secret key file: %s", err.Error())
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			log.Printf("error closing file: %s", err.Error())
		}
	}()

	pubKeyFile, err := os.Open("jwtRS256.key.pub")
	if err != nil {
		log.Fatalf("error opening public key file: %s", err.Error())
	}
	defer func() {
		if err := pubKeyFile.Close(); err != nil {
			log.Printf("error closing file: %s", err.Error())
		}
	}()

	server.AuthClient, err = auth.InitJwtAuth(keyFile, pubKeyFile)
	if err != nil {
		log.Printf("error initializing JWT authentication: %s", err.Error())
	} else {
		server.AuthRoutes()
		log.Println("\tJWT Auth initialized")
	}
}
