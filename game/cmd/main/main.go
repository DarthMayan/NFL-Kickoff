package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"kickoff.com/game/internal/data"
	"kickoff.com/game/internal/database"
	"kickoff.com/game/internal/models"
	pb "kickoff.com/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const serviceName = "game"

type GameService struct {
	pb.UnimplementedGameServiceServer
}

func main() {
	var grpcPort int
	flag.IntVar(&grpcPort, "grpc-port", 9082, "gRPC server port")
	flag.Parse()

	log.Printf("Starting Game Service - gRPC:%d", grpcPort)
	log.Printf("Service discovery: Kubernetes DNS")

	// Conectar a la base de datos
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL database")

	// Cargar equipos NFL (solo si no existen)
	loadNFLTeams()

	// Cargar juegos de ejemplo
	loadSampleGames()

	// Inicializar servicio
	gameService := &GameService{}

	// Crear listener para gRPC
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear servidor gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterGameServiceServer(grpcServer, gameService)

	// Registrar health check
	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("✅ gRPC server listening on :%d", grpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	database.Close()
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

// ========================================
// Helper Functions
// ========================================

func loadNFLTeams() {
	log.Println("Loading NFL teams...")
	for _, teamData := range data.NFLTeams {
		var existing models.Team
		result := database.DB.Where("id = ?", teamData.ID).First(&existing)

		if result.Error != nil {
			if err := database.DB.Create(&teamData).Error; err != nil {
				log.Printf("Error creating team %s: %v", teamData.ID, err)
			} else {
				log.Printf("Created team: %s", teamData.Name)
			}
		}
	}
	log.Println("✅ NFL teams loaded")
}

func loadSampleGames() {
	log.Println("Loading sample games...")
	sampleGames := []models.Game{
		{
			ID:         "game_1",
			Week:       1,
			Season:     2024,
			HomeTeamID: "KC",
			AwayTeamID: "SF",
			GameTime:   time.Now().Add(24 * time.Hour),
			Status:     models.GameStatusScheduled,
			HomeScore:  0,
			AwayScore:  0,
		},
		{
			ID:         "game_2",
			Week:       1,
			Season:     2024,
			HomeTeamID: "BUF",
			AwayTeamID: "DAL",
			GameTime:   time.Now().Add(48 * time.Hour),
			Status:     models.GameStatusScheduled,
			HomeScore:  0,
			AwayScore:  0,
		},
		{
			ID:         "game_3",
			Week:       1,
			Season:     2024,
			HomeTeamID: "PHI",
			AwayTeamID: "NYG",
			GameTime:   time.Now().Add(-2 * time.Hour),
			Status:     models.GameStatusCompleted,
			HomeScore:  28,
			AwayScore:  14,
			WinnerTeamID: "PHI",
		},
		{
			ID:         "game_4",
			Week:       1,
			Season:     2024,
			HomeTeamID: "GB",
			AwayTeamID: "CHI",
			GameTime:   time.Now(),
			Status:     models.GameStatusLive,
			HomeScore:  21,
			AwayScore:  10,
		},
	}

	for _, game := range sampleGames {
		var existing models.Game
		result := database.DB.Where("id = ?", game.ID).First(&existing)
		if result.Error != nil {
			if err := database.DB.Create(&game).Error; err != nil {
				log.Printf("Error creating game %s: %v", game.ID, err)
			} else {
				log.Printf("Created game: %s", game.ID)
			}
		}
	}
	log.Println("✅ Sample games loaded")
}

func modelTeamToProto(team models.Team) *pb.Team {
	return &pb.Team{
		Id:           team.ID,
		Name:         team.Name,
		City:         team.City,
		Abbreviation: team.ID,
		Conference:   conferenceToProto(team.Conference),
		Division:     divisionToProto(team.Division),
		LogoUrl:      "",
		Stadium:      team.Stadium,
	}
}

func conferenceToProto(conf models.Conference) pb.Conference {
	switch conf {
	case models.ConferenceAFC:
		return pb.Conference_CONFERENCE_AFC
	case models.ConferenceNFC:
		return pb.Conference_CONFERENCE_NFC
	default:
		return pb.Conference_CONFERENCE_UNSPECIFIED
	}
}

func conferenceFromProto(conf pb.Conference) models.Conference {
	switch conf {
	case pb.Conference_CONFERENCE_AFC:
		return models.ConferenceAFC
	case pb.Conference_CONFERENCE_NFC:
		return models.ConferenceNFC
	default:
		return ""
	}
}

func divisionToProto(div models.Division) pb.Division {
	switch div {
	case models.DivisionAFCEast:
		return pb.Division_DIVISION_AFC_EAST
	case models.DivisionAFCNorth:
		return pb.Division_DIVISION_AFC_NORTH
	case models.DivisionAFCSouth:
		return pb.Division_DIVISION_AFC_SOUTH
	case models.DivisionAFCWest:
		return pb.Division_DIVISION_AFC_WEST
	case models.DivisionNFCEast:
		return pb.Division_DIVISION_NFC_EAST
	case models.DivisionNFCNorth:
		return pb.Division_DIVISION_NFC_NORTH
	case models.DivisionNFCSouth:
		return pb.Division_DIVISION_NFC_SOUTH
	case models.DivisionNFCWest:
		return pb.Division_DIVISION_NFC_WEST
	default:
		return pb.Division_DIVISION_UNSPECIFIED
	}
}

func divisionFromProto(div pb.Division) models.Division {
	switch div {
	case pb.Division_DIVISION_AFC_EAST:
		return models.DivisionAFCEast
	case pb.Division_DIVISION_AFC_NORTH:
		return models.DivisionAFCNorth
	case pb.Division_DIVISION_AFC_SOUTH:
		return models.DivisionAFCSouth
	case pb.Division_DIVISION_AFC_WEST:
		return models.DivisionAFCWest
	case pb.Division_DIVISION_NFC_EAST:
		return models.DivisionNFCEast
	case pb.Division_DIVISION_NFC_NORTH:
		return models.DivisionNFCNorth
	case pb.Division_DIVISION_NFC_SOUTH:
		return models.DivisionNFCSouth
	case pb.Division_DIVISION_NFC_WEST:
		return models.DivisionNFCWest
	default:
		return ""
	}
}

func modelGameToProto(game models.Game) *pb.Game {
	protoGame := &pb.Game{
		Id:          game.ID,
		HomeTeamId:  game.HomeTeamID,
		AwayTeamId:  game.AwayTeamID,
		Week:        int32(game.Week),
		Status:      gameStatusToProto(game.Status),
		HomeScore:   int32(game.HomeScore),
		AwayScore:   int32(game.AwayScore),
		ScheduledAt: timestamppb.New(game.GameTime),
	}

	if !game.CreatedAt.IsZero() {
		protoGame.StartedAt = timestamppb.New(game.CreatedAt)
	}

	if game.Status == models.GameStatusCompleted && !game.UpdatedAt.IsZero() {
		protoGame.CompletedAt = timestamppb.New(game.UpdatedAt)
	}

	return protoGame
}

func gameStatusToProto(status models.GameStatus) pb.GameStatus {
	switch status {
	case models.GameStatusScheduled:
		return pb.GameStatus_GAME_STATUS_SCHEDULED
	case models.GameStatusLive:
		return pb.GameStatus_GAME_STATUS_IN_PROGRESS
	case models.GameStatusCompleted:
		return pb.GameStatus_GAME_STATUS_COMPLETED
	case models.GameStatusPostponed:
		return pb.GameStatus_GAME_STATUS_POSTPONED
	case models.GameStatusCanceled:
		return pb.GameStatus_GAME_STATUS_CANCELED
	default:
		return pb.GameStatus_GAME_STATUS_UNSPECIFIED
	}
}

func gameStatusFromProto(status pb.GameStatus) models.GameStatus {
	switch status {
	case pb.GameStatus_GAME_STATUS_SCHEDULED:
		return models.GameStatusScheduled
	case pb.GameStatus_GAME_STATUS_IN_PROGRESS:
		return models.GameStatusLive
	case pb.GameStatus_GAME_STATUS_COMPLETED:
		return models.GameStatusCompleted
	case pb.GameStatus_GAME_STATUS_POSTPONED:
		return models.GameStatusPostponed
	case pb.GameStatus_GAME_STATUS_CANCELED:
		return models.GameStatusCanceled
	default:
		return models.GameStatusScheduled
	}
}

// ========================================
// gRPC Handlers - Team Operations
// ========================================

func (gs *GameService) GetAllTeams(ctx context.Context, req *pb.GetAllTeamsRequest) (*pb.GetAllTeamsResponse, error) {
	var teams []models.Team
	if err := database.DB.Find(&teams).Error; err != nil {
		log.Printf("Error fetching teams: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch teams: %v", err)
	}

	var pbTeams []*pb.Team
	for _, team := range teams {
		pbTeams = append(pbTeams, modelTeamToProto(team))
	}

	return &pb.GetAllTeamsResponse{
		Teams: pbTeams,
		Total: int32(len(pbTeams)),
	}, nil
}

func (gs *GameService) GetTeamByID(ctx context.Context, req *pb.GetTeamByIDRequest) (*pb.GetTeamByIDResponse, error) {
	if req.TeamId == "" {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}

	teamID := strings.ToUpper(req.TeamId)
	var team models.Team
	if err := database.DB.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Team not found")
	}

	return &pb.GetTeamByIDResponse{
		Team: modelTeamToProto(team),
	}, nil
}

func (gs *GameService) GetTeamsByConference(ctx context.Context, req *pb.GetTeamsByConferenceRequest) (*pb.GetTeamsByConferenceResponse, error) {
	if req.Conference == pb.Conference_CONFERENCE_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "conference is required")
	}

	targetConference := conferenceFromProto(req.Conference)
	var teams []models.Team
	if err := database.DB.Where("conference = ?", targetConference).Find(&teams).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch teams: %v", err)
	}

	var pbTeams []*pb.Team
	for _, team := range teams {
		pbTeams = append(pbTeams, modelTeamToProto(team))
	}

	return &pb.GetTeamsByConferenceResponse{
		Teams:      pbTeams,
		Total:      int32(len(pbTeams)),
		Conference: req.Conference,
	}, nil
}

func (gs *GameService) GetTeamsByDivision(ctx context.Context, req *pb.GetTeamsByDivisionRequest) (*pb.GetTeamsByDivisionResponse, error) {
	if req.Division == pb.Division_DIVISION_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "division is required")
	}

	targetDivision := divisionFromProto(req.Division)
	var teams []models.Team
	if err := database.DB.Where("division = ?", targetDivision).Find(&teams).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch teams: %v", err)
	}

	var pbTeams []*pb.Team
	for _, team := range teams {
		pbTeams = append(pbTeams, modelTeamToProto(team))
	}

	return &pb.GetTeamsByDivisionResponse{
		Teams:    pbTeams,
		Total:    int32(len(pbTeams)),
		Division: req.Division,
	}, nil
}

// ========================================
// gRPC Handlers - Game Operations
// ========================================

func (gs *GameService) GetAllGames(ctx context.Context, req *pb.GetAllGamesRequest) (*pb.GetAllGamesResponse, error) {
	var games []models.Game
	if err := database.DB.Find(&games).Error; err != nil {
		log.Printf("Error fetching games: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch games: %v", err)
	}

	var pbGames []*pb.Game
	for _, game := range games {
		pbGames = append(pbGames, modelGameToProto(game))
	}

	return &pb.GetAllGamesResponse{
		Games: pbGames,
		Total: int32(len(pbGames)),
	}, nil
}

func (gs *GameService) GetGameByID(ctx context.Context, req *pb.GetGameByIDRequest) (*pb.GetGameByIDResponse, error) {
	if req.GameId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id is required")
	}

	var game models.Game
	if err := database.DB.Where("id = ?", req.GameId).First(&game).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Game not found")
	}

	return &pb.GetGameByIDResponse{
		Game: modelGameToProto(game),
	}, nil
}

func (gs *GameService) GetGamesByWeek(ctx context.Context, req *pb.GetGamesByWeekRequest) (*pb.GetGamesByWeekResponse, error) {
	if req.Week < 1 || req.Week > 18 {
		return nil, status.Error(codes.InvalidArgument, "week must be between 1 and 18")
	}

	var games []models.Game
	if err := database.DB.Where("week = ?", req.Week).Find(&games).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch games: %v", err)
	}

	var pbGames []*pb.Game
	for _, game := range games {
		pbGames = append(pbGames, modelGameToProto(game))
	}

	return &pb.GetGamesByWeekResponse{
		Games: pbGames,
		Total: int32(len(pbGames)),
		Week:  req.Week,
	}, nil
}

func (gs *GameService) GetGamesByTeam(ctx context.Context, req *pb.GetGamesByTeamRequest) (*pb.GetGamesByTeamResponse, error) {
	if req.TeamId == "" {
		return nil, status.Error(codes.InvalidArgument, "team_id is required")
	}

	teamID := strings.ToUpper(req.TeamId)

	// Verificar que el equipo existe
	var team models.Team
	if err := database.DB.Where("id = ?", teamID).First(&team).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Team not found")
	}

	var games []models.Game
	if err := database.DB.Where("home_team_id = ? OR away_team_id = ?", teamID, teamID).Find(&games).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch games: %v", err)
	}

	var pbGames []*pb.Game
	for _, game := range games {
		pbGames = append(pbGames, modelGameToProto(game))
	}

	return &pb.GetGamesByTeamResponse{
		Games:  pbGames,
		Total:  int32(len(pbGames)),
		TeamId: teamID,
	}, nil
}

func (gs *GameService) GetGamesByStatus(ctx context.Context, req *pb.GetGamesByStatusRequest) (*pb.GetGamesByStatusResponse, error) {
	if req.Status == pb.GameStatus_GAME_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	targetStatus := gameStatusFromProto(req.Status)
	var games []models.Game
	if err := database.DB.Where("status = ?", targetStatus).Find(&games).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch games: %v", err)
	}

	var pbGames []*pb.Game
	for _, game := range games {
		pbGames = append(pbGames, modelGameToProto(game))
	}

	return &pb.GetGamesByStatusResponse{
		Games:  pbGames,
		Total:  int32(len(pbGames)),
		Status: req.Status,
	}, nil
}

// ========================================
// gRPC Handlers - Game Management
// ========================================

func (gs *GameService) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	if req.HomeTeamId == "" || req.AwayTeamId == "" {
		return nil, status.Error(codes.InvalidArgument, "home_team_id and away_team_id are required")
	}

	if req.Week < 1 || req.Week > 18 {
		return nil, status.Error(codes.InvalidArgument, "week must be between 1 and 18")
	}

	homeTeamID := strings.ToUpper(req.HomeTeamId)
	awayTeamID := strings.ToUpper(req.AwayTeamId)

	// Verificar que los equipos existen
	var homeTeam, awayTeam models.Team
	if err := database.DB.Where("id = ?", homeTeamID).First(&homeTeam).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Home team not found")
	}
	if err := database.DB.Where("id = ?", awayTeamID).First(&awayTeam).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Away team not found")
	}

	if homeTeamID == awayTeamID {
		return nil, status.Error(codes.InvalidArgument, "A team cannot play against itself")
	}

	// Generar ID único
	var count int64
	database.DB.Model(&models.Game{}).Count(&count)
	gameID := fmt.Sprintf("game_%d", count+1)

	scheduledAt := time.Now().Add(24 * time.Hour)
	if req.ScheduledAt != nil {
		scheduledAt = req.ScheduledAt.AsTime()
	}

	game := models.Game{
		ID:         gameID,
		Week:       int(req.Week),
		Season:     2024,
		HomeTeamID: homeTeamID,
		AwayTeamID: awayTeamID,
		GameTime:   scheduledAt,
		Status:     models.GameStatusScheduled,
		HomeScore:  0,
		AwayScore:  0,
	}

	if err := database.DB.Create(&game).Error; err != nil {
		log.Printf("Error creating game: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create game: %v", err)
	}

	return &pb.CreateGameResponse{
		Game:    modelGameToProto(game),
		Message: "Game created successfully",
	}, nil
}

func (gs *GameService) UpdateGameScore(ctx context.Context, req *pb.UpdateGameScoreRequest) (*pb.UpdateGameScoreResponse, error) {
	if req.GameId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id is required")
	}

	var game models.Game
	if err := database.DB.Where("id = ?", req.GameId).First(&game).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Game not found")
	}

	updates := map[string]interface{}{
		"home_score": int(req.HomeScore),
		"away_score": int(req.AwayScore),
	}

	if err := database.DB.Model(&game).Updates(updates).Error; err != nil {
		log.Printf("Error updating game score: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update game score: %v", err)
	}

	// Recargar juego actualizado
	database.DB.Where("id = ?", req.GameId).First(&game)

	return &pb.UpdateGameScoreResponse{
		Game:    modelGameToProto(game),
		Message: "Game score updated successfully",
	}, nil
}

func (gs *GameService) UpdateGameStatus(ctx context.Context, req *pb.UpdateGameStatusRequest) (*pb.UpdateGameStatusResponse, error) {
	if req.GameId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id is required")
	}

	if req.Status == pb.GameStatus_GAME_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	var game models.Game
	if err := database.DB.Where("id = ?", req.GameId).First(&game).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Game not found")
	}

	newStatus := gameStatusFromProto(req.Status)
	updates := map[string]interface{}{
		"status": newStatus,
	}

	// Si el juego se completa, determinar el ganador
	if newStatus == models.GameStatusCompleted {
		if game.HomeScore > game.AwayScore {
			updates["winner_team_id"] = game.HomeTeamID
		} else if game.AwayScore > game.HomeScore {
			updates["winner_team_id"] = game.AwayTeamID
		}
	}

	if err := database.DB.Model(&game).Updates(updates).Error; err != nil {
		log.Printf("Error updating game status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update game status: %v", err)
	}

	// Recargar juego actualizado
	database.DB.Where("id = ?", req.GameId).First(&game)

	return &pb.UpdateGameStatusResponse{
		Game:    modelGameToProto(game),
		Message: "Game status updated successfully",
	}, nil
}
