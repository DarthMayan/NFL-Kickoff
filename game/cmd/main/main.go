package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
)

const serviceName = "game"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()

	log.Printf("Starting game service on port %d", port)

	// Crear conexión con Consul
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)

	// Registrar servicio en Consul
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}

	// Goroutine para reportar estado de salud cada segundo
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Asegurar que se desregistre al terminar
	defer registry.Deregister(ctx, instanceID, serviceName)

	// Endpoint básico de prueba
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Game service is healthy"))
	})

	http.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"teams": [
				{"id": "KC", "name": "Kansas City Chiefs", "conference": "AFC"},
				{"id": "SF", "name": "San Francisco 49ers", "conference": "NFC"},
				{"id": "BUF", "name": "Buffalo Bills", "conference": "AFC"},
				{"id": "DAL", "name": "Dallas Cowboys", "conference": "NFC"}
			]
		}`))
	})

	http.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"games": [
				{"id": "1", "homeTeam": "KC", "awayTeam": "SF", "week": 1, "status": "scheduled"},
				{"id": "2", "homeTeam": "BUF", "awayTeam": "DAL", "week": 1, "status": "scheduled"}
			]
		}`))
	})

	log.Printf("Game service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
