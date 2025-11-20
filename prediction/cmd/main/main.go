package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
	"kickoff.com/pkg/models"
)

const serviceName = "prediction"

type PredictionService struct {
	predictions map[string]models.Prediction
	counter     int
	registry    discovery.Registry
}

// Estructura para mantener compatibilidad con el formato anterior
type OldPrediction struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	GameID          string    `json:"gameId"`
	PredictedWinner string    `json:"predictedWinner"`
	CreatedAt       time.Time `json:"createdAt"`
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8083, "API handler port")
	flag.Parse()

	log.Printf("Starting prediction service on port %d", port)

	// Crear conexión con Consul
	consulAddr := os.Getenv("CONSUL_ADDRESS")
	if consulAddr == "" {
		log.Fatal("CONSUL_ADDRESS environment variable is required")
	}
	registry, err := consul.NewRegistry(consulAddr)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)

	// Registrar servicio en Consul con el nombre del contenedor
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("prediction-service:%d", port)); err != nil {
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

	// Inicializar servicio
	predictionService := &PredictionService{
		predictions: make(map[string]models.Prediction),
		counter:     1,
		registry:    registry,
	}

	// === ENDPOINTS EXISTENTES (mantener funcionando) ===
	http.HandleFunc("/health", predictionService.healthHandler)
	http.HandleFunc("/predictions", predictionService.predictionsHandlerOld)           // Endpoint original
	http.HandleFunc("/predictions/user/", predictionService.userPredictionsHandlerOld) // Endpoint original

	// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
	http.HandleFunc("/v2/predictions", predictionService.predictionsHandlerNew)           // CRUD completo
	http.HandleFunc("/v2/predictions/", predictionService.predictionByIDHandler)          // Predicción por ID
	http.HandleFunc("/v2/predictions/user/", predictionService.userPredictionsHandlerNew) // Predicciones de usuario
	http.HandleFunc("/v2/predictions/game/", predictionService.gamePredictionsHandler)    // Predicciones de juego
	http.HandleFunc("/v2/predictions/week/", predictionService.weekPredictionsHandler)    // Predicciones por semana

	log.Printf("Prediction service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func (ps *PredictionService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Prediction service is healthy"))
}

// === ENDPOINTS ORIGINALES (mantener compatibilidad) ===
func (ps *PredictionService) predictionsHandlerOld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Convertir a formato anterior para compatibilidad
		var oldPredictions []OldPrediction
		for _, prediction := range ps.predictions {
			oldPred := OldPrediction{
				ID:              prediction.ID,
				UserID:          prediction.UserID,
				GameID:          prediction.GameID,
				PredictedWinner: prediction.PredictedWinnerID,
				CreatedAt:       prediction.CreatedAt,
			}
			oldPredictions = append(oldPredictions, oldPred)
		}

		response := map[string]interface{}{
			"predictions": oldPredictions,
			"total":       len(oldPredictions),
		}
		json.NewEncoder(w).Encode(response)

	case "POST":
		// Aceptar formato anterior
		var req struct {
			UserID          string `json:"userId"`
			GameID          string `json:"gameId"`
			PredictedWinner string `json:"predictedWinner"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validar campos requeridos
		if req.UserID == "" || req.GameID == "" || req.PredictedWinner == "" {
			http.Error(w, "Missing required fields: userId, gameId, predictedWinner", http.StatusBadRequest)
			return
		}

		// Crear predicción usando modelo nuevo
		prediction := models.Prediction{
			ID:                fmt.Sprintf("pred_%d", ps.counter),
			UserID:            req.UserID,
			GameID:            req.GameID,
			PredictedWinnerID: req.PredictedWinner,
			Status:            models.PredictionStatusPending,
			Points:            0,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		ps.counter++
		ps.predictions[prediction.ID] = prediction

		// Responder con formato anterior
		oldPred := OldPrediction{
			ID:              prediction.ID,
			UserID:          prediction.UserID,
			GameID:          prediction.GameID,
			PredictedWinner: prediction.PredictedWinnerID,
			CreatedAt:       prediction.CreatedAt,
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(oldPred)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ps *PredictionService) userPredictionsHandlerOld(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/predictions/user/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Filtrar predicciones por usuario y convertir a formato anterior
	var oldPredictions []OldPrediction
	for _, prediction := range ps.predictions {
		if prediction.UserID == userID {
			oldPred := OldPrediction{
				ID:              prediction.ID,
				UserID:          prediction.UserID,
				GameID:          prediction.GameID,
				PredictedWinner: prediction.PredictedWinnerID,
				CreatedAt:       prediction.CreatedAt,
			}
			oldPredictions = append(oldPredictions, oldPred)
		}
	}

	response := map[string]interface{}{
		"userId":      userID,
		"predictions": oldPredictions,
		"total":       len(oldPredictions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
func (ps *PredictionService) predictionsHandlerNew(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		var predictions []models.Prediction
		for _, prediction := range ps.predictions {
			predictions = append(predictions, prediction)
		}

		response := models.PredictionsResponse{
			Predictions: predictions,
			Total:       len(predictions),
		}
		json.NewEncoder(w).Encode(response)

	case "POST":
		var req models.CreatePredictionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Validar campos requeridos
		if req.UserID == "" || req.GameID == "" || req.PredictedWinnerID == "" {
			http.Error(w, "Missing required fields: userId, gameId, predictedWinnerId", http.StatusBadRequest)
			return
		}

		// Verificar que no exista predicción para este usuario y juego
		for _, existing := range ps.predictions {
			if existing.UserID == req.UserID && existing.GameID == req.GameID {
				http.Error(w, "Prediction already exists for this game", http.StatusConflict)
				return
			}
		}

		// Validar que el equipo predicho participe en el juego (simulado)
		if !ps.validateTeamInGame(req.GameID, req.PredictedWinnerID) {
			http.Error(w, "Predicted team does not participate in this game", http.StatusBadRequest)
			return
		}

		// Crear predicción
		prediction := models.Prediction{
			ID:                fmt.Sprintf("pred_%d", ps.counter),
			UserID:            req.UserID,
			GameID:            req.GameID,
			PredictedWinnerID: req.PredictedWinnerID,
			Status:            models.PredictionStatusPending,
			Points:            0,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		ps.counter++
		ps.predictions[prediction.ID] = prediction

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(prediction)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ps *PredictionService) predictionByIDHandler(w http.ResponseWriter, r *http.Request) {
	predictionID := strings.TrimPrefix(r.URL.Path, "/v2/predictions/")

	switch r.Method {
	case "GET":
		prediction, exists := ps.predictions[predictionID]
		if !exists {
			http.Error(w, "Prediction not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prediction)

	case "DELETE":
		prediction, exists := ps.predictions[predictionID]
		if !exists {
			http.Error(w, "Prediction not found", http.StatusNotFound)
			return
		}

		// Solo permitir eliminar predicciones pendientes
		if prediction.Status != models.PredictionStatusPending {
			http.Error(w, "Cannot delete non-pending prediction", http.StatusBadRequest)
			return
		}

		delete(ps.predictions, predictionID)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ps *PredictionService) userPredictionsHandlerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/v2/predictions/user/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var userPredictions []models.Prediction
	correct, incorrect, pending := 0, 0, 0

	for _, prediction := range ps.predictions {
		if prediction.UserID == userID {
			userPredictions = append(userPredictions, prediction)
			switch prediction.Status {
			case models.PredictionStatusCorrect:
				correct++
			case models.PredictionStatusIncorrect:
				incorrect++
			case models.PredictionStatusPending:
				pending++
			}
		}
	}

	var percentage float64
	if len(userPredictions) > 0 {
		percentage = float64(correct) / float64(len(userPredictions)) * 100
	}

	response := models.UserPredictionsResponse{
		UserID:      userID,
		Predictions: userPredictions,
		Total:       len(userPredictions),
		Correct:     correct,
		Incorrect:   incorrect,
		Pending:     pending,
		Percentage:  percentage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ps *PredictionService) gamePredictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	gameID := strings.TrimPrefix(r.URL.Path, "/v2/predictions/game/")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	var gamePredictions []models.Prediction
	for _, prediction := range ps.predictions {
		if prediction.GameID == gameID {
			gamePredictions = append(gamePredictions, prediction)
		}
	}

	response := map[string]interface{}{
		"gameId":      gameID,
		"predictions": gamePredictions,
		"total":       len(gamePredictions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ps *PredictionService) weekPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	week := strings.TrimPrefix(r.URL.Path, "/v2/predictions/week/")
	if week == "" {
		http.Error(w, "Week number is required", http.StatusBadRequest)
		return
	}

	// Para simplificar, asumimos que los gameIDs contienen la semana
	// En implementación real, esto requeriría consultar Game Service
	var weekPredictions []models.Prediction
	for _, prediction := range ps.predictions {
		// Simulación simple: juegos 1-16 son semana 1, etc.
		if (prediction.GameID == "1" || prediction.GameID == "2") && week == "1" {
			weekPredictions = append(weekPredictions, prediction)
		}
	}

	response := map[string]interface{}{
		"week":        week,
		"predictions": weekPredictions,
		"total":       len(weekPredictions),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función auxiliar para validar que el equipo participe en el juego
func (ps *PredictionService) validateTeamInGame(gameID, teamID string) bool {
	// Simulación simple basada en los juegos de ejemplo
	validTeams := map[string][]string{
		"1": {"KC", "SF"},   // Juego 1: KC vs SF
		"2": {"BUF", "DAL"}, // Juego 2: BUF vs DAL
	}

	teams, exists := validTeams[gameID]
	if !exists {
		return false
	}

	for _, team := range teams {
		if team == teamID {
			return true
		}
	}
	return false
}
