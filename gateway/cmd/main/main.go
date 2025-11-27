package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	pb "kickoff.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const serviceName = "gateway"

type Gateway struct {
	userClient        pb.UserServiceClient
	gameClient        pb.GameServiceClient
	predictionClient  pb.PredictionServiceClient
	leaderboardClient pb.LeaderboardServiceClient
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "API handler port")
	flag.Parse()

	log.Printf("Starting Gateway service on port %d", port)
	log.Printf("Using Kubernetes DNS for service discovery")

	gateway := &Gateway{}

	// Inicializar conexiones gRPC a los servicios
	if err := gateway.initGRPCClients(); err != nil {
		log.Fatalf("Failed to initialize gRPC clients: %v", err)
	}

	// Endpoints del Gateway
	http.HandleFunc("/", gateway.frontendHandler)
	http.HandleFunc("/health", gateway.corsMiddleware(gateway.healthHandler))
	http.HandleFunc("/api/users", gateway.corsMiddleware(gateway.usersHandler))
	http.HandleFunc("/api/teams", gateway.corsMiddleware(gateway.teamsHandler))
	// Support both listing and single-game lookup: /api/games and /api/games/{id}
	http.HandleFunc("/api/games", gateway.corsMiddleware(gateway.gamesHandler))
	http.HandleFunc("/api/games/", gateway.corsMiddleware(gateway.gamesHandler))
	http.HandleFunc("/api/predictions", gateway.corsMiddleware(gateway.predictionsHandler))
	http.HandleFunc("/api/predictions/user/", gateway.corsMiddleware(gateway.userPredictionsHandler))
	http.HandleFunc("/api/leaderboard", gateway.corsMiddleware(gateway.leaderboardHandler))
	http.HandleFunc("/api/user-stats/", gateway.corsMiddleware(gateway.userStatsHandler))

	log.Printf("Gateway service listening on :%d", port)
	log.Println("✅ All gRPC clients initialized")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}

// ========================================
// gRPC Client Initialization
// ========================================

func (g *Gateway) initGRPCClients() error {
	var err error

	// Connect to User Service via gRPC
	userConn, err := grpc.NewClient("user-service:9081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to user service: %v", err)
	}
	g.userClient = pb.NewUserServiceClient(userConn)
	log.Println("✅ Connected to User Service gRPC (user-service:9081)")

	// Connect to Game Service via gRPC
	gameConn, err := grpc.NewClient("game-service:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to game service: %v", err)
	}
	g.gameClient = pb.NewGameServiceClient(gameConn)
	log.Println("✅ Connected to Game Service gRPC (game-service:9082)")

	// Connect to Prediction Service via gRPC
	predictionConn, err := grpc.NewClient("prediction-service:9083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to prediction service: %v", err)
	}
	g.predictionClient = pb.NewPredictionServiceClient(predictionConn)
	log.Println("✅ Connected to Prediction Service gRPC (prediction-service:9083)")

	// Connect to Leaderboard Service via gRPC
	leaderboardConn, err := grpc.NewClient("leaderboard-service:9084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to leaderboard service: %v", err)
	}
	g.leaderboardClient = pb.NewLeaderboardServiceClient(leaderboardConn)
	log.Println("✅ Connected to Leaderboard Service gRPC (leaderboard-service:9084)")

	return nil
}

// ========================================
// CORS Middleware
// ========================================

func (g *Gateway) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the actual handler
		next(w, r)
	}
}

// ========================================
// HTTP Handlers (usando gRPC internamente)
// ========================================

func (g *Gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Gateway service is healthy"))
}

func (g *Gateway) frontendHandler(w http.ResponseWriter, r *http.Request) {
	// Solo servir HTML en la raíz
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	html := `<!doctype html>
<html lang="es">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>Kickoff - NFL Predictions</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      min-height: 100vh;
      padding: 20px;
      color: #333;
    }
    .container { max-width: 1200px; margin: 0 auto; }
    header {
      background: rgba(255,255,255,0.95);
      padding: 24px;
      border-radius: 8px;
      margin-bottom: 24px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    }
    header h1 { font-size: 2em; margin-bottom: 8px; }
    .status { display: flex; gap: 16px; align-items: center; margin-top: 12px; }
    .status-pill { display: inline-flex; align-items: center; gap: 6px; padding: 8px 12px; border-radius: 20px; font-size: 0.9em; }
    .status-ok { background: #d4edda; color: #155724; }
    .status-err { background: #f8d7da; color: #721c24; }

    .sections { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
    .section {
      background: white;
      border-radius: 8px;
      padding: 20px;
      box-shadow: 0 2px 8px rgba(0,0,0,0.1);
      overflow: hidden;
    }
    .section h2 { font-size: 1.4em; margin-bottom: 16px; color: #667eea; border-bottom: 2px solid #667eea; padding-bottom: 12px; }
    .section-content { max-height: 500px; overflow-y: auto; }

    .team-card, .game-card, .user-card, .pred-card {
      background: #f8f9fa;
      padding: 12px;
      margin-bottom: 12px;
      border-radius: 6px;
      border-left: 4px solid #667eea;
    }
    .team-card h3, .game-card h3, .user-card h3, .pred-card h3 {
      font-size: 1em;
      margin-bottom: 6px;
    }
    .team-card p, .game-card p, .user-card p, .pred-card p {
      font-size: 0.9em;
      color: #666;
      margin: 4px 0;
    }

    .game-status {
      display: inline-block;
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 0.85em;
      font-weight: bold;
    }
    .status-1 { background: #cfe2ff; color: #084298; }
    .status-2 { background: #f8d7da; color: #842029; }
    .status-3 { background: #d1e7dd; color: #0f5132; }

    .loading { text-align: center; color: #999; font-style: italic; }
    .error { background: #f8d7da; color: #721c24; padding: 12px; border-radius: 6px; margin-bottom: 12px; }
    .empty { text-align: center; color: #999; padding: 20px; font-style: italic; }

    footer {
      text-align: center;
      color: white;
      margin-top: 40px;
      font-size: 0.9em;
    }

    @media (max-width: 768px) {
      header h1 { font-size: 1.4em; }
      .sections { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <h1>🏈 Kickoff - NFL Predictions</h1>
      <div class="status">
        <span>Gateway Status:</span>
        <span class="status-pill status-err" id="healthStatus">Checking...</span>
      </div>
    </header>

    <div class="sections">
      <div class="section">
        <h2>🏟️ Equipos NFL</h2>
        <div class="section-content" id="teamsContainer">
          <div class="loading">Cargando equipos...</div>
        </div>
      </div>

      <div class="section">
        <h2>🎮 Juegos</h2>
        <div class="section-content" id="gamesContainer">
          <div class="loading">Cargando juegos...</div>
        </div>
      </div>

      <div class="section">
        <h2>👥 Usuarios</h2>
        <div class="section-content" id="usersContainer">
          <div class="loading">Cargando usuarios...</div>
        </div>
      </div>

      <div class="section">
        <h2>🏆 Leaderboard</h2>
        <div class="section-content" id="leaderboardContainer">
          <div class="loading">Cargando leaderboard...</div>
        </div>
      </div>

      <div class="section">
        <h2>🔮 Predicciones</h2>
        <div class="section-content" id="predictionsContainer">
          <div class="loading">Cargando predicciones...</div>
        </div>
      </div>

      <div class="section">
        <h2>📊 Estadísticas</h2>
        <div class="section-content" id="statsContainer">
          <div class="loading">Cargando información...</div>
        </div>
      </div>
    </div>

    <footer>
      <p>Kickoff NFL Predictions • Kubernetes + gRPC + PostgreSQL</p>
    </footer>
  </div>

  <script>
    const API_BASE = window.location.origin;

    async function apiCall(endpoint) {
      try {
        const res = await fetch(API_BASE + endpoint);
        if (!res.ok) throw new Error('HTTP ' + res.status);
        return await res.json();
      } catch (err) {
        console.error('Error fetching ' + endpoint + ':', err);
        return null;
      }
    }

    async function checkHealth() {
      try {
        const res = await fetch(API_BASE + '/health');
        const statusEl = document.getElementById('healthStatus');
        if (res.ok) {
          statusEl.classList.remove('status-err');
          statusEl.classList.add('status-ok');
          statusEl.textContent = '✅ Online';
        } else {
          statusEl.classList.remove('status-ok');
          statusEl.classList.add('status-err');
          statusEl.textContent = '❌ Offline';
        }
      } catch(e) {
        const statusEl = document.getElementById('healthStatus');
        statusEl.classList.remove('status-ok');
        statusEl.classList.add('status-err');
        statusEl.textContent = '❌ Error';
      }
    }

    async function loadTeams() {
      const data = await apiCall('/api/teams');
      const el = document.getElementById('teamsContainer');
      if (!data || !data.teams || data.teams.length === 0) {
        el.innerHTML = '<div class="empty">No hay equipos disponibles</div>';
        return;
      }
      el.innerHTML = data.teams.slice(0, 8).map(t =>
        '<div class="team-card"><h3>' + t.name + '</h3>' +
        '<p><strong>' + t.id + '</strong> • ' + t.city + '</p>' +
        '<p>' + t.stadium + '</p></div>'
      ).join('');
    }

    async function loadGames() {
      const data = await apiCall('/api/games');
      const el = document.getElementById('gamesContainer');
      if (!data || !data.games || data.games.length === 0) {
        el.innerHTML = '<div class="empty">No hay juegos disponibles</div>';
        return;
      }
      el.innerHTML = data.games.map(g => {
        const statusText = g.status === 1 ? 'Programado' : g.status === 2 ? 'En Vivo' : 'Finalizado';
        return '<div class="game-card"><h3>' + g.home_team_id + ' vs ' + g.away_team_id + '</h3>' +
          '<p><strong>Semana ' + g.week + '</strong></p>' +
          '<p>Score: <strong>' + (g.home_score || 0) + '-' + (g.away_score || 0) + '</strong></p>' +
          '<p><span class="game-status status-' + g.status + '">' + statusText + '</span></p></div>';
      }).join('');
    }

    async function loadUsers() {
      const data = await apiCall('/api/users');
      const el = document.getElementById('usersContainer');
      if (!data || !data.users || data.users.length === 0) {
        el.innerHTML = '<div class="empty">No hay usuarios registrados</div>';
        return;
      }
      el.innerHTML = data.users.slice(0, 10).map(u =>
        '<div class="user-card"><h3>' + (u.full_name || u.username) + '</h3>' +
        '<p>@' + u.username + '</p>' +
        '<p>' + u.email + '</p></div>'
      ).join('');
    }

    async function loadLeaderboard() {
      const data = await apiCall('/api/leaderboard');
      const el = document.getElementById('leaderboardContainer');
      if (!data || !data.leaderboard || data.leaderboard.length === 0) {
        el.innerHTML = '<div class="empty">El leaderboard está vacío</div>';
        return;
      }
      el.innerHTML = data.leaderboard.slice(0, 10).map((u, i) =>
        '<div class="user-card"><h3>#' + (i + 1) + ' User ' + u.user_id + '</h3>' +
        '<p><strong>' + (u.correct_picks || 0) + '</strong> de <strong>' + (u.total_picks || 0) + '</strong> correctas</p>' +
        '<p>Precisión: ' + (u.percentage || 0).toFixed(1) + '%</p></div>'
      ).join('');
    }

    async function loadPredictions() {
      const data = await apiCall('/api/predictions');
      const el = document.getElementById('predictionsContainer');
      if (!data || !data.predictions || data.predictions.length === 0) {
        el.innerHTML = '<div class="empty">No hay predicciones</div>';
        return;
      }
      el.innerHTML = data.predictions.slice(0, 10).map(p =>
        '<div class="pred-card"><h3>Juego ' + p.game_id + '</h3>' +
        '<p>Usuario: <strong>' + p.user_id + '</strong></p>' +
        '<p>Predicción: <strong>' + p.predicted_winner_id + '</strong></p>' +
        '<p>Puntos: ' + (p.points || 0) + '</p></div>'
      ).join('');
    }

    async function loadStats() {
      const el = document.getElementById('statsContainer');
      const teams = await apiCall('/api/teams');
      const games = await apiCall('/api/games');
      const users = await apiCall('/api/users');
      const preds = await apiCall('/api/predictions');
      const lb = await apiCall('/api/leaderboard');

      el.innerHTML =
        '<div class="user-card"><p><strong>Equipos:</strong> ' + (teams?.total || 0) + '</p></div>' +
        '<div class="user-card"><p><strong>Juegos:</strong> ' + (games?.total || 0) + '</p></div>' +
        '<div class="user-card"><p><strong>Usuarios:</strong> ' + (users?.total || 0) + '</p></div>' +
        '<div class="user-card"><p><strong>Predicciones:</strong> ' + (preds?.total || 0) + '</p></div>' +
        '<div class="user-card"><p><strong>En Ranking:</strong> ' + (lb?.total_users || 0) + '</p></div>';
    }

    async function init() {
      await checkHealth();
      await Promise.all([
        loadTeams(),
        loadGames(),
        loadUsers(),
        loadLeaderboard(),
        loadPredictions(),
        loadStats()
      ]);
    }

    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', init);
    } else {
      init();
    }

    setInterval(init, 30000);
  </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (g *Gateway) usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := g.userClient.GetAllUsers(ctx, &pb.GetAllUsersRequest{
			Page:     1,
			PageSize: 100,
		})
		if err != nil {
			log.Printf("Error calling user service: %v", err)
			http.Error(w, "Error calling user service", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": resp.Users,
			"total": resp.Total,
		})

	case "POST":
		var reqBody struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			FullName string `json:"fullName"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		reqBody.Username = strings.TrimSpace(reqBody.Username)
		reqBody.Email = strings.TrimSpace(reqBody.Email)
		reqBody.FullName = strings.TrimSpace(reqBody.FullName)

		if reqBody.Username == "" || reqBody.Email == "" {
			http.Error(w, "username and email are required", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := g.userClient.CreateUser(ctx, &pb.CreateUserRequest{
			Username: reqBody.Username,
			Email:    reqBody.Email,
			FullName: reqBody.FullName,
		})
		if err != nil {
			log.Printf("Error creating user: %v", err)
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"user":    resp.User,
			"message": resp.Message,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) predictionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if r.Method == "GET" {
		// Llamar al Prediction Service via gRPC
		resp, err := g.predictionClient.GetAllPredictions(ctx, &pb.GetAllPredictionsRequest{})
		if err != nil {
			log.Printf("Error calling prediction service: %v", err)
			http.Error(w, "Error calling prediction service", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"predictions": resp.Predictions,
			"total":       resp.Total,
		})

	} else if r.Method == "POST" {
		// Leer el body del request
		var reqBody struct {
			UserID          string `json:"userId"`
			GameID          string `json:"gameId"`
			PredictedWinner string `json:"predictedWinner"`
		}

		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Crear predicción via gRPC
		resp, err := g.predictionClient.CreatePrediction(ctx, &pb.CreatePredictionRequest{
			UserId:            reqBody.UserID,
			GameId:            reqBody.GameID,
			PredictedWinnerId: reqBody.PredictedWinner,
		})
		if err != nil {
			log.Printf("Error creating prediction: %v", err)
			http.Error(w, "Error creating prediction", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"prediction": resp.Prediction,
			"message":    resp.Message,
		})

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (g *Gateway) userPredictionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Extraer userID de la URL
	userID := strings.TrimPrefix(r.URL.Path, "/api/predictions/user/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Llamar al Prediction Service via gRPC
	resp, err := g.predictionClient.GetUserPredictions(ctx, &pb.GetUserPredictionsRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("Error getting user predictions: %v", err)
		http.Error(w, "Error getting user predictions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"predictions": resp.Predictions,
		"userId":      resp.UserId,
		"total":       resp.Total,
	})
}

func (g *Gateway) leaderboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Llamar al Leaderboard Service via gRPC
	resp, err := g.leaderboardClient.GetLeaderboard(ctx, &pb.GetLeaderboardRequest{})
	if err != nil {
		log.Printf("Error getting leaderboard: %v", err)
		http.Error(w, "Error getting leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"leaderboard":   resp.Leaderboard,
		"totalUsers":    resp.TotalUsers,
		"gamesFinished": resp.GamesFinished,
	})
}

func (g *Gateway) userStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Extraer userID de la URL
	userID := strings.TrimPrefix(r.URL.Path, "/api/user-stats/")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Llamar al Leaderboard Service via gRPC
	resp, err := g.leaderboardClient.GetUserStats(ctx, &pb.GetUserStatsRequest{
		UserId: userID,
	})
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		http.Error(w, "Error getting user stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userStats":   resp.UserStats,
		"predictions": resp.Predictions,
	})
}

func (g *Gateway) teamsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Llamar al Game Service via gRPC
	resp, err := g.gameClient.GetAllTeams(ctx, &pb.GetAllTeamsRequest{})
	if err != nil {
		log.Printf("Error getting teams: %v", err)
		http.Error(w, "Error getting teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"teams": resp.Teams,
		"total": resp.Total,
	})
}

func (g *Gateway) gamesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Determine if the request is for a single game (path: /api/games/{id})
	// or for the list (/api/games or /api/games/)
	raw := strings.TrimPrefix(r.URL.Path, "/api/games")
	id := strings.Trim(raw, "/")

	if id != "" {
		// Single game lookup
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		resp, err := g.gameClient.GetGameByID(ctx, &pb.GetGameByIDRequest{GameId: id})
		if err != nil {
			log.Printf("Error getting game by id '%s': %v", id, err)
			http.Error(w, "Game not found or error calling game service", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"game": resp.Game,
		})
		return
	}

	// No ID provided: return all games
	resp, err := g.gameClient.GetAllGames(ctx, &pb.GetAllGamesRequest{})
	if err != nil {
		log.Printf("Error getting games: %v", err)
		http.Error(w, "Error getting games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"games": resp.Games,
		"total": resp.Total,
	})
}
