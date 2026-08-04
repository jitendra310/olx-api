package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/jitendra310/olx-api/internal/config"
	"github.com/jitendra310/olx-api/internal/handlers"
)

func main() {
	cfg := config.MustLoad()

	fmt.Println("starting olx server")

	//make own router
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handlers.Health)

	srv := http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 30,
		IdleTimeout:  time.Second * 60,
	}

	// err := http.ListenAndServe(":8090", mux) //we can use this also for temp
	log.Printf("Server is listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server faild: %v", err)
	}
}
