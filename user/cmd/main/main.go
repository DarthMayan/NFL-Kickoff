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
	"kickoff.com/user/internal/controller"
	httphandler "kickoff.com/user/internal/handler"
	"kickoff.com/user/internal/repository/memory"
)

const serviceName = "user"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Printf("Starting user service on port %d", port)

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

	// Goroutine para reportar estado de salud
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	// Crear las capas del servicio
	repo := memory.New()
	ctrl := controller.New(repo)
	handler := httphandler.New(ctrl)

	// Registrar endpoints
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User service is healthy"))
	})
	http.HandleFunc("/user", handler.GetUser)
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler.GetAllUsers(w, r)
		} else if r.Method == "POST" {
			handler.CreateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("User service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
