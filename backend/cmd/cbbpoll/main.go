package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/cors"
	"golang.org/x/crypto/acme/autocert"

	_ "github.com/r-cbb/cbbpoll/docs"
	"github.com/r-cbb/cbbpoll/internal/app"
	"github.com/r-cbb/cbbpoll/internal/auth"
	"github.com/r-cbb/cbbpoll/internal/db/sqlite"
	"github.com/r-cbb/cbbpoll/internal/server"
)

func main() {
	log.SetOutput(os.Stdout)
	log.Println("Initializing server...")

	srv := server.NewServer()
	var err error

	// Setup Database connection
	db, err := sqlite.NewClient("/data/cbbpoll.db")
	if err != nil {
		log.Fatal(err.Error())
		panic(err.Error())
	}
	log.Println("\tSqlite Client initialized")

	// Setup service layer
	srv.App = app.NewPollService(db)
	srv.App.Admins = append(srv.App.Admins, "Concision", "einsteins_haircut")

	// Setup JWT Auth
	setupAuth(srv)

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
	srv.RedditClient = server.NewRedditClient("https://oauth.reddit.com/api/v1")

	// Enable CORS for Swagger-UI
	c := cors.New(cors.Options{
		Debug:          false,
		AllowedHeaders: []string{"*"},
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{},
		MaxAge:         1000,
	})

	// TODO: config
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "http://localhost:8000"
	}
	srv.SetHost(host)

	var useSSL bool
	httpsEnabled := os.Getenv("HTTPS_ENABLED")
	if httpsEnabled == "1" || strings.ToLower(httpsEnabled) == "true" {
		useSSL = true
	}

	var handler http.Handler
	handler = c.Handler(srv)

	if useSSL {
		dataDir := "/data/ssl_cache"
		hostPolicy := func(ctx context.Context, host string) error {
			allowedHost := "api.cbbpoll.com"
			if host == allowedHost {
				return nil
			}
			return fmt.Errorf("acme/autocert: only %s host is allowed", allowedHost)
		}

		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache(dataDir),
		}
		httpsSrv := makeHTTPServer(handler, "443")
		httpsSrv.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}

		d1 := serverListen(httpsSrv, true)

		httpSrv := makeHTTPServer(m.HTTPHandler(nil), "80")
		d2 := serverListen(httpSrv, false)
		<-d1
		<-d2
		log.Println("Done")

	} else {
		httpSrv := makeHTTPServer(handler, port)
		done := serverListen(httpSrv, false)
		<-done
		log.Println("Done")
	}

	err = db.Close()
	if err != nil {
		log.Printf("Error closing db: %s", err.Error())
	}

	os.Exit(0)
}

func serverListen(s *http.Server, tls bool) chan bool {
	done := make(chan bool, 1)
	signalled := make(chan os.Signal, 1)
	signal.Notify(signalled, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
	go func() {
		if tls {
			log.Println(s.ListenAndServeTLS("", ""))
		} else {
			log.Println(s.ListenAndServe())
		}
	}()
	log.Printf("Serving %s", s.Addr)
	go func() {
		<-signalled
		log.Println("Server stopping")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		_ = s.Shutdown(ctx)
		log.Println("Server stopped")
		close(done)
	}()

	return done
}

func makeHTTPServer(handler http.Handler, port string) *http.Server {
	return &http.Server{
		Handler:      handler,
		Addr:         fmt.Sprintf(":%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func setupAuth(server *server.Server) {
	keyFile, err := os.Open("/data/jwtRS256.key")
	if err != nil {
		log.Fatalf("error opening secret key file: %s", err.Error())
	}
	defer func() {
		if err := keyFile.Close(); err != nil {
			log.Printf("error closing file: %s", err.Error())
		}
	}()

	pubKeyFile, err := os.Open("/data/jwtRS256.key.pub")
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
