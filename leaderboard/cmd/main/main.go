package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
)

const serviceName = "leaderboard"

type UserScore struct {
	UserID       string  `json:"userId"`
	CorrectPicks int     `json:"correctPicks"`
	TotalPicks   int     `json:"totalPicks"`
	Percentage   float64 `json:"percentage"`
	Rank         int     `json:"rank"`
}

type Game struct {
	ID       string `json:"id"`
	HomeTeam string `json:"homeTeam"`
	AwayTeam string `json:"awayTeam"`
	Winner   string `json:"winner,omitempty"` // Para juegos finalizados
	Status   string `json:"status"`
}

type Prediction struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	GameID          string    `json:"gameId"`
	PredictedWinner string    `json:"predictedWinner"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Leaderboard struct {
	registry discovery.Registry
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8084, "API handler port")
	flag.Parse()

	log.Printf("Starting leaderboard service on port %d", port)

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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("leaderboard-service:%d", port)); err != nil {
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

	leaderboard := &Leaderboard{registry: registry}

	// Endpoints
	http.HandleFunc("/health", leaderboard.healthHandler)
	http.HandleFunc("/leaderboard", leaderboard.leaderboardHandler)
	http.HandleFunc("/user-stats/", leaderboard.userStatsHandler)

	log.Printf("Leaderboard service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func (l *Leaderboard) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Leaderboard service is healthy"))
}

func (l *Leaderboard) leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Obtener todas las predicciones del Prediction Service
	predictions, err := l.getAllPredictions()
	if err != nil {
		http.Error(w, "Error fetching predictions", http.StatusInternalServerError)
		return
	}

	// Para simplificar, vamos a simular que algunos juegos han terminado
	// En una implementación real, esto vendría del Game Service
	finishedGames := map[string]string{
		"1": "KC",  // Kansas City ganó
		"2": "BUF", // Buffalo ganó
	}

	// Calcular estadísticas por usuario
	userStats := make(map[string]*UserScore)

	for _, prediction := range predictions {
		if userStats[prediction.UserID] == nil {
			userStats[prediction.UserID] = &UserScore{
				UserID: prediction.UserID,
			}
		}

		stats := userStats[prediction.UserID]
		stats.TotalPicks++

		// Verificar si la predicción fue correcta
		if winner, gameFinished := finishedGames[prediction.GameID]; gameFinished {
			if prediction.PredictedWinner == winner {
				stats.CorrectPicks++
			}
		}
	}

	// Calcular porcentajes y convertir a slice para ordenar
	var leaderboard []UserScore
	for _, stats := range userStats {
		if stats.TotalPicks > 0 {
			stats.Percentage = float64(stats.CorrectPicks) / float64(stats.TotalPicks) * 100
		}
		leaderboard = append(leaderboard, *stats)
	}

	// Ordenar por porcentaje (descendente)
	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Percentage > leaderboard[j].Percentage
	})

	// Asignar rankings
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	response := map[string]interface{}{
		"leaderboard":   leaderboard,
		"totalUsers":    len(leaderboard),
		"gamesFinished": len(finishedGames),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (l *Leaderboard) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer userID de la URL
	userID := r.URL.Path[len("/user-stats/"):]
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Obtener predicciones del usuario específico
	predictions, err := l.getUserPredictions(userID)
	if err != nil {
		http.Error(w, "Error fetching user predictions", http.StatusInternalServerError)
		return
	}

	// Simular resultados de juegos
	finishedGames := map[string]string{
		"1": "KC",
		"2": "BUF",
	}

	stats := UserScore{UserID: userID}
	var detailedPredictions []map[string]interface{}

	for _, prediction := range predictions {
		stats.TotalPicks++

		predictionDetail := map[string]interface{}{
			"gameId":          prediction.GameID,
			"predictedWinner": prediction.PredictedWinner,
			"createdAt":       prediction.CreatedAt,
		}

		if winner, gameFinished := finishedGames[prediction.GameID]; gameFinished {
			predictionDetail["actualWinner"] = winner
			predictionDetail["correct"] = prediction.PredictedWinner == winner
			if prediction.PredictedWinner == winner {
				stats.CorrectPicks++
			}
		} else {
			predictionDetail["gameStatus"] = "pending"
		}

		detailedPredictions = append(detailedPredictions, predictionDetail)
	}

	if stats.TotalPicks > 0 {
		stats.Percentage = float64(stats.CorrectPicks) / float64(stats.TotalPicks) * 100
	}

	response := map[string]interface{}{
		"userStats":   stats,
		"predictions": detailedPredictions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (l *Leaderboard) getAllPredictions() ([]Prediction, error) {
	// Buscar el Prediction Service en Consul
	addresses, err := l.registry.ServiceAddress(context.Background(), "prediction")
	if err != nil {
		return nil, err
	}

	// Llamar al endpoint de predicciones
	url := fmt.Sprintf("http://%s/predictions", addresses[0])
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Predictions []Prediction `json:"predictions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Predictions, nil
}

func (l *Leaderboard) getUserPredictions(userID string) ([]Prediction, error) {
	// Buscar el Prediction Service en Consul
	addresses, err := l.registry.ServiceAddress(context.Background(), "prediction")
	if err != nil {
		return nil, err
	}

	// Llamar al endpoint de predicciones del usuario
	url := fmt.Sprintf("http://%s/predictions/user/%s", addresses[0], userID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Predictions []Prediction `json:"predictions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Predictions, nil
}
