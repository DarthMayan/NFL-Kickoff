# Kickoff - NFL Prediction Game

A distributed microservices application for NFL game predictions and leaderboards built with Go and Consul for service discovery.

## Architecture

Kickoff follows a microservices architecture with the following services:

- **User Service** (port 8081) - User management and authentication
- **Game Service** (port 8082) - NFL teams and games management
- **Prediction Service** (port 8083) - User predictions for games
- **Leaderboard Service** (port 8084) - Rankings and statistics
- **Gateway Service** (port 8080) - API Gateway and main entry point
- **Consul** (port 8500) - Service discovery and health checking

## Prerequisites

- Go 1.21 or higher
- Docker Desktop
- Git

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd kickoff
```

2. Start Consul:
```bash
./run-consul.bat
```

3. Install dependencies:
```bash
go mod tidy
```

## Running the Services

Start each service in a separate terminal:

```bash
# Terminal 1 - User Service
go run user/cmd/main/main.go

# Terminal 2 - Game Service  
go run game/cmd/main/main.go

# Terminal 3 - Prediction Service
go run prediction/cmd/main/main.go

# Terminal 4 - Leaderboard Service
go run leaderboard/cmd/main/main.go

# Terminal 5 - Gateway Service
go run gateway/cmd/main/main.go
```

## API Endpoints

### Gateway (Port 8080)
- `GET /api/users` - Get all users
- `POST /api/users` - Create a new user
- `GET /api/teams` - Get all NFL teams
- `GET /api/games` - Get all games
- `GET /api/predictions` - Get all predictions
- `POST /api/predictions` - Create a prediction
- `GET /api/leaderboard` - Get current leaderboard

### Direct Service Access
Each service also exposes endpoints directly on their respective ports for development and debugging.

## Example Usage

1. Create a user:
```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"username": "john_doe", "email": "john@example.com", "fullName": "John Doe"}'
```

2. Make a prediction:
```bash
curl -X POST http://localhost:8080/api/predictions \
  -H "Content-Type: application/json" \
  -d '{"userId": "user_john_doe", "gameId": "1", "predictedWinnerId": "KC"}'
```

3. Check leaderboard:
```bash
curl http://localhost:8080/api/leaderboard
```

## Service Discovery

All services automatically register with Consul and can discover each other dynamically. Check the Consul UI at http://localhost:8500 to see all registered services.

## Development

The project follows a clean architecture pattern with:
- Repository pattern for data access
- Controller layer for business logic  
- HTTP handlers for API endpoints
- Service-specific models and interfaces

## Tech Stack

- **Language**: Go
- **Service Discovery**: Consul
- **Architecture**: Microservices
- **Storage**: In-memory (development)
- **Containerization**: Docker

## Project Status

This is a learning project for distributed systems development. Current features include basic CRUD operations, service discovery, and inter-service communication.