package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/seriousben/dev-identity-provider/server"
)

const (
	envServerPort       = "SERVER_PORT"
	envServerRemoteAddr = "SERVER_REMOTE_ADDR"
)

func main() {
	var (
		serverPort       = os.Getenv(envServerPort)
		serverRemoteAddr = os.Getenv(envServerRemoteAddr)
	)

	if serverPort == "" {
		log.Fatalf("missing %s environment variable", envServerPort)
	}
	if serverRemoteAddr == "" {
		log.Fatalf("missing %s environment variable", envServerRemoteAddr)
	}

	h := server.New(serverRemoteAddr)

	srv := &http.Server{
		Handler:      h,
		Addr:         fmt.Sprintf(":%s", serverPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Println("Starting server")

	log.Fatal(srv.ListenAndServe())
}
