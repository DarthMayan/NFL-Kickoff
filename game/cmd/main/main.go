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

	"kickoff.com/game/internal/data"
	"kickoff.com/pkg/discovery"
	"kickoff.com/pkg/discovery/consul"
	"kickoff.com/pkg/models"
)

const serviceName = "game"

type GameService struct {
	teams map[string]models.Team
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "API handler port")
	flag.Parse()

	log.Printf("Starting game service on port %d", port)

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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("game-service:%d", port)); err != nil {
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

	// Inicializar servicio con datos de equipos
	gameService := &GameService{
		teams: make(map[string]models.Team),
	}

	// Cargar todos los equipos NFL
	for _, team := range data.NFLTeams {
		gameService.teams[team.ID] = team
	}

	// === ENDPOINTS EXISTENTES (mantener funcionando) ===
	http.HandleFunc("/health", gameService.healthHandler)
	http.HandleFunc("/teams", gameService.teamsHandlerOld) // Endpoint original
	http.HandleFunc("/games", gameService.gamesHandlerOld) // Endpoint original

	// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
	http.HandleFunc("/v2/teams", gameService.teamsHandlerNew)                      // Todos los equipos
	http.HandleFunc("/v2/teams/", gameService.teamByIDHandler)                     // Equipo por ID
	http.HandleFunc("/v2/teams/conference/", gameService.teamsByConferenceHandler) // Por conferencia
	http.HandleFunc("/v2/teams/division/", gameService.teamsByDivisionHandler)     // Por división
	http.HandleFunc("/v2/games", gameService.gamesHandlerNew)                      // Todos los juegos

	log.Printf("Game service listening on :%d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

func (gs *GameService) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Game service is healthy"))
}

// === ENDPOINTS ORIGINALES (NO TOCAR) ===
func (gs *GameService) teamsHandlerOld(w http.ResponseWriter, r *http.Request) {
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
}

func (gs *GameService) gamesHandlerOld(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"games": [
			{"id": "1", "homeTeam": "KC", "awayTeam": "SF", "week": 1, "status": "scheduled"},
			{"id": "2", "homeTeam": "BUF", "awayTeam": "DAL", "week": 1, "status": "scheduled"}
		]
	}`))
}

// === NUEVOS ENDPOINTS CON MODELOS ROBUSTOS ===
func (gs *GameService) teamsHandlerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var teams []models.Team
	for _, team := range gs.teams {
		teams = append(teams, team)
	}

	response := models.TeamsResponse{
		Teams: teams,
		Total: len(teams),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GameService) teamByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer team ID de la URL (/v2/teams/{teamID})
	teamID := strings.TrimPrefix(r.URL.Path, "/v2/teams/")
	teamID = strings.ToUpper(teamID)

	team, exists := gs.teams[teamID]
	if !exists {
		http.Error(w, "Team not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (gs *GameService) teamsByConferenceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer conference de la URL (/v2/teams/conference/{conference})
	conference := strings.TrimPrefix(r.URL.Path, "/v2/teams/conference/")
	conference = strings.ToUpper(conference)

	if conference != "AFC" && conference != "NFC" {
		http.Error(w, "Invalid conference. Must be AFC or NFC", http.StatusBadRequest)
		return
	}

	var teams []models.Team
	targetConference := models.Conference(conference)

	for _, team := range gs.teams {
		if team.Conference == targetConference {
			teams = append(teams, team)
		}
	}

	response := models.TeamsResponse{
		Teams: teams,
		Total: len(teams),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (gs *GameService) teamsByDivisionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extraer division de la URL (/v2/teams/division/{division})
	divisionPath := strings.TrimPrefix(r.URL.Path, "/v2/teams/division/")

	// Convertir path a división válida
	var targetDivision models.Division
	switch strings.ToLower(divisionPath) {
	case "afc-east":
		targetDivision = models.DivisionAFCEast
	case "afc-north":
		targetDivision = models.DivisionAFCNorth
	case "afc-south":
		targetDivision = models.DivisionAFCSouth
	case "afc-west":
		targetDivision = models.DivisionAFCWest
	case "nfc-east":
		targetDivision = models.DivisionNFCEast
	case "nfc-north":
		targetDivision = models.DivisionNFCNorth
	case "nfc-south":
		targetDivision = models.DivisionNFCSouth
	case "nfc-west":
		targetDivision = models.DivisionNFCWest
	default:
		http.Error(w, "Invalid division", http.StatusBadRequest)
		return
	}

	var teams []models.Team
	for _, team := range gs.teams {
		if team.Division == targetDivision {
			teams = append(teams, team)
		}
	}

	response := models.TeamsResponse{
		Teams: teams,
		Total: len(teams),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// === GAMES HANDLERS ===
func (gs *GameService) gamesHandlerNew(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Datos de ejemplo de juegos NFL
	games := []map[string]interface{}{
		{"id": "1", "homeTeam": "KC", "awayTeam": "SF", "week": 1, "status": "scheduled", "homeScore": 0, "awayScore": 0},
		{"id": "2", "homeTeam": "BUF", "awayTeam": "DAL", "week": 1, "status": "scheduled", "homeScore": 0, "awayScore": 0},
		{"id": "3", "homeTeam": "PHI", "awayTeam": "NYG", "week": 1, "status": "completed", "homeScore": 28, "awayScore": 14},
		{"id": "4", "homeTeam": "GB", "awayTeam": "CHI", "week": 1, "status": "in_progress", "homeScore": 21, "awayScore": 10},
	}

	response := map[string]interface{}{
		"games": games,
		"total": len(games),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
