package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
)

const serviceName = "prediction"

type Prediction struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	GameID          string    `json:"gameId"`
	PredictedWinner string    `json:"predictedWinner"`
	CreatedAt       time.Time `json:"createdAt"`
}

// Simple in-memory storage for now
var predictions = make(map[string]Prediction)
var predictionCounter = 1

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API handler port")
	flag.Parse()

	log.Printf("Starting prediction service on port %d", port)

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

	// Endpoints
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/predictions", predictionsHandler)
	http.HandleFunc("/predictions/user/", userPredictionsHandler)

	log.Printf("Prediction service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Prediction service is healthy"))
}

func predictionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Obtener todas las predicciones
		var allPredictions []Prediction
		for _, prediction := range predictions {
			allPredictions = append(allPredictions, prediction)
		}

		response := map[string]interface{}{
			"predictions": allPredictions,
			"total":       len(allPredictions),
		}

		json.NewEncoder(w).Encode(response)

	case "POST":
		// Crear nueva predicción
		var newPrediction Prediction
		if err := json.NewDecoder(r.Body).Decode(&newPrediction); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Generar ID y timestamp
		newPrediction.ID = fmt.Sprintf("pred_%d", predictionCounter)
		newPrediction.CreatedAt = time.Now()
		predictionCounter++

		// Validar campos requeridos
		if newPrediction.UserID == "" || newPrediction.GameID == "" || newPrediction.PredictedWinner == "" {
			http.Error(w, "Missing required fields: userId, gameId, predictedWinner", http.StatusBadRequest)
			return
		}

		// Guardar predicción
		predictions[newPrediction.ID] = newPrediction

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newPrediction)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func userPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer userID de la URL (formato: /predictions/user/{userID})
	userID := r.URL.Path[len("/predictions/user/"):]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Filtrar predicciones por usuario
	var userPredictions []Prediction
	for _, prediction := range predictions {
		if prediction.UserID == userID {
			userPredictions = append(userPredictions, prediction)
		}
	}

	response := map[string]interface{}{
		"userId":      userID,
		"predictions": userPredictions,
		"total":       len(userPredictions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
