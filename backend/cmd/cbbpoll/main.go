package main

import (
	"fmt"
	"github.com/r-cbb/cbbpoll/backend/internal/app"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("hello")

	server := app.NewServer()

	srv := &http.Server{
		Handler: server.Handler(),
		Addr:    "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}