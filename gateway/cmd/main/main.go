package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
)

const serviceName = "gateway"

type Gateway struct {
	registry discovery.Registry
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "API handler port")
	flag.Parse()

	log.Printf("Starting gateway service on port %d", port)

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

	gateway := &Gateway{registry: registry}

	// Endpoints del Gateway
	http.HandleFunc("/health", gateway.healthHandler)
	http.HandleFunc("/api/users", gateway.usersHandler)
	http.HandleFunc("/api/teams", gateway.teamsHandler)
	http.HandleFunc("/api/games", gateway.gamesHandler)
	http.HandleFunc("/api/predictions", gateway.predictionsHandler)
	http.HandleFunc("/api/predictions/user/", gateway.userPredictionsHandler)
	http.HandleFunc("/api/leaderboard", gateway.leaderboardHandler)
	http.HandleFunc("/api/user-stats/", gateway.userStatsHandler)

	log.Printf("Gateway service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func (g *Gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Gateway service is healthy"))
}

func (g *Gateway) usersHandler(w http.ResponseWriter, r *http.Request) {
	// Buscar el servicio user en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "user")
	if err != nil {
		http.Error(w, "User service not available", http.StatusServiceUnavailable)
		return
	}

	// Usar la primera dirección disponible
	userServiceURL := fmt.Sprintf("http://%s/users", addresses[0])

	// Llamar al servicio user
	resp, err := http.Get(userServiceURL)
	if err != nil {
		http.Error(w, "Error calling user service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Reenviar la respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *Gateway) predictionsHandler(w http.ResponseWriter, r *http.Request) {
	// Buscar el servicio prediction en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "prediction")
	if err != nil {
		http.Error(w, "Prediction service not available", http.StatusServiceUnavailable)
		return
	}

	// Construir URL y reenviar la solicitud
	predictionServiceURL := fmt.Sprintf("http://%s/predictions", addresses[0])

	if r.Method == "GET" {
		resp, err := http.Get(predictionServiceURL)
		if err != nil {
			http.Error(w, "Error calling prediction service", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else if r.Method == "POST" {
		// Reenviar POST request con body
		req, err := http.NewRequest("POST", predictionServiceURL, r.Body)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error calling prediction service", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) userPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer userID de la URL
	userID := r.URL.Path[len("/api/predictions/user/"):]

	// Buscar el servicio prediction en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "prediction")
	if err != nil {
		http.Error(w, "Prediction service not available", http.StatusServiceUnavailable)
		return
	}

	// Llamar al servicio prediction
	predictionServiceURL := fmt.Sprintf("http://%s/predictions/user/%s", addresses[0], userID)
	resp, err := http.Get(predictionServiceURL)
	if err != nil {
		http.Error(w, "Error calling prediction service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *Gateway) leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Buscar el servicio leaderboard en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "leaderboard")
	if err != nil {
		http.Error(w, "Leaderboard service not available", http.StatusServiceUnavailable)
		return
	}

	// Llamar al servicio leaderboard
	leaderboardServiceURL := fmt.Sprintf("http://%s/leaderboard", addresses[0])
	resp, err := http.Get(leaderboardServiceURL)
	if err != nil {
		http.Error(w, "Error calling leaderboard service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *Gateway) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer userID de la URL
	userID := r.URL.Path[len("/api/user-stats/"):]

	// Buscar el servicio leaderboard en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "leaderboard")
	if err != nil {
		http.Error(w, "Leaderboard service not available", http.StatusServiceUnavailable)
		return
	}

	// Llamar al servicio leaderboard
	leaderboardServiceURL := fmt.Sprintf("http://%s/user-stats/%s", addresses[0], userID)
	resp, err := http.Get(leaderboardServiceURL)
	if err != nil {
		http.Error(w, "Error calling leaderboard service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *Gateway) teamsHandler(w http.ResponseWriter, r *http.Request) {
	// Buscar el servicio game en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "game")
	if err != nil {
		http.Error(w, "Game service not available", http.StatusServiceUnavailable)
		return
	}

	// Usar la primera dirección disponible
	gameServiceURL := fmt.Sprintf("http://%s/teams", addresses[0])

	// Llamar al servicio game
	resp, err := http.Get(gameServiceURL)
	if err != nil {
		http.Error(w, "Error calling game service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Reenviar la respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (g *Gateway) gamesHandler(w http.ResponseWriter, r *http.Request) {
	// Buscar el servicio game en Consul
	addresses, err := g.registry.ServiceAddress(context.Background(), "game")
	if err != nil {
		http.Error(w, "Game service not available", http.StatusServiceUnavailable)
		return
	}

	// Usar la primera dirección disponible
	gameServiceURL := fmt.Sprintf("http://%s/games", addresses[0])

	// Llamar al servicio game
	resp, err := http.Get(gameServiceURL)
	if err != nil {
		http.Error(w, "Error calling game service", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Reenviar la respuesta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
