package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"

	"kickoff.com/leaderboard/internal/database"
	"kickoff.com/leaderboard/internal/models"
	pb "kickoff.com/proto"
)

const serviceName = "leaderboard"

type LeaderboardService struct {
	pb.UnimplementedLeaderboardServiceServer
}

func main() {
	var grpcPort int
	flag.IntVar(&grpcPort, "grpc-port", 9084, "gRPC server port")
	flag.Parse()

	log.Printf("Starting Leaderboard Service - gRPC:%d", grpcPort)
	log.Printf("Service discovery: Kubernetes DNS")

	// Conectar a la base de datos
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL database")

	leaderboardService := &LeaderboardService{}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterLeaderboardServiceServer(grpcServer, leaderboardService)

	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("✅ gRPC server listening on :%d", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")
	database.Close()
	grpcServer.GracefulStop()
}

func (ls *LeaderboardService) GetLeaderboard(ctx context.Context, req *pb.GetLeaderboardRequest) (*pb.GetLeaderboardResponse, error) {
	var userStats []models.UserStats
	query := database.DB.Order("total_points DESC, correct_predictions DESC")

	if req.Limit > 0 {
		query = query.Limit(int(req.Limit))
	}
	if req.Offset > 0 {
		query = query.Offset(int(req.Offset))
	}

	if err := query.Find(&userStats).Error; err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch leaderboard: %v", err)
	}

	// Actualizar rangos
	for i := range userStats {
		userStats[i].Rank = i + 1 + int(req.Offset)
		database.DB.Model(&userStats[i]).Update("rank", userStats[i].Rank)
	}

	var pbLeaderboard []*pb.UserScore
	for _, stats := range userStats {
		pbLeaderboard = append(pbLeaderboard, &pb.UserScore{
			UserId:      stats.UserID,
			CorrectPicks: int32(stats.CorrectPredictions),
			TotalPicks:  int32(stats.TotalPredictions),
			Percentage:  calculatePercentage(stats.CorrectPredictions, stats.TotalPredictions),
			Rank:        int32(stats.Rank),
		})
	}

	// Contar total de usuarios
	var totalUsers int64
	database.DB.Model(&models.UserStats{}).Count(&totalUsers)

	return &pb.GetLeaderboardResponse{
		Leaderboard:   pbLeaderboard,
		TotalUsers:    int32(totalUsers),
		GamesFinished: 0, // TODO: obtener de game service
	}, nil
}

func (ls *LeaderboardService) GetUserStats(ctx context.Context, req *pb.GetUserStatsRequest) (*pb.GetUserStatsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var userStats models.UserStats
	result := database.DB.Where("user_id = ?", req.UserId).First(&userStats)

	if result.Error != nil {
		// Si no existe, crear un registro inicial
		userStats = models.UserStats{
			ID:                 fmt.Sprintf("stats_%s", req.UserId),
			UserID:             req.UserId,
			TotalPredictions:   0,
			CorrectPredictions: 0,
			WrongPredictions:   0,
			TotalPoints:        0,
			Rank:               0,
		}
		if err := database.DB.Create(&userStats).Error; err != nil {
			log.Printf("Error creating user stats: %v", err)
			return nil, status.Errorf(codes.Internal, "failed to create user stats: %v", err)
		}
	}

	return &pb.GetUserStatsResponse{
		UserStats: &pb.UserScore{
			UserId:      userStats.UserID,
			CorrectPicks: int32(userStats.CorrectPredictions),
			TotalPicks:  int32(userStats.TotalPredictions),
			Percentage:  calculatePercentage(userStats.CorrectPredictions, userStats.TotalPredictions),
			Rank:        int32(userStats.Rank),
		},
		Predictions:      []*pb.PredictionDetail{}, // TODO: obtener de prediction service
		TotalPredictions: int32(userStats.TotalPredictions),
	}, nil
}

func (ls *LeaderboardService) GetTopUsers(ctx context.Context, req *pb.GetTopUsersRequest) (*pb.GetTopUsersResponse, error) {
	limit := int(req.TopN)
	if limit <= 0 {
		limit = 10
	}

	var userStats []models.UserStats
	if err := database.DB.Order("total_points DESC, correct_predictions DESC").Limit(limit).Find(&userStats).Error; err != nil {
		log.Printf("Error fetching top users: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch top users: %v", err)
	}

	// Actualizar rangos
	for i := range userStats {
		userStats[i].Rank = i + 1
		database.DB.Model(&userStats[i]).Update("rank", userStats[i].Rank)
	}

	var pbPlayers []*pb.UserScore
	for _, stats := range userStats {
		pbPlayers = append(pbPlayers, &pb.UserScore{
			UserId:      stats.UserID,
			CorrectPicks: int32(stats.CorrectPredictions),
			TotalPicks:  int32(stats.TotalPredictions),
			Percentage:  calculatePercentage(stats.CorrectPredictions, stats.TotalPredictions),
			Rank:        int32(stats.Rank),
		})
	}

	return &pb.GetTopUsersResponse{
		TopUsers: pbPlayers,
		Total:    int32(len(pbPlayers)),
	}, nil
}

func (ls *LeaderboardService) GetUserRank(ctx context.Context, req *pb.GetUserRankRequest) (*pb.GetUserRankResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var userStats models.UserStats
	if err := database.DB.Where("user_id = ?", req.UserId).First(&userStats).Error; err != nil {
		return nil, status.Error(codes.NotFound, "User stats not found")
	}

	// Contar cuántos usuarios tienen mejor puntaje
	var betterCount int64
	database.DB.Model(&models.UserStats{}).
		Where("total_points > ? OR (total_points = ? AND correct_predictions > ?)",
			userStats.TotalPoints, userStats.TotalPoints, userStats.CorrectPredictions).
		Count(&betterCount)

	rank := int(betterCount) + 1
	userStats.Rank = rank
	database.DB.Model(&userStats).Update("rank", rank)

	// Contar total de usuarios
	var totalUsers int64
	database.DB.Model(&models.UserStats{}).Count(&totalUsers)

	return &pb.GetUserRankResponse{
		UserScore: &pb.UserScore{
			UserId:      userStats.UserID,
			CorrectPicks: int32(userStats.CorrectPredictions),
			TotalPicks:  int32(userStats.TotalPredictions),
			Percentage:  calculatePercentage(userStats.CorrectPredictions, userStats.TotalPredictions),
			Rank:        int32(rank),
		},
		Rank:       int32(rank),
		TotalUsers: int32(totalUsers),
	}, nil
}

func (ls *LeaderboardService) RecalculateLeaderboard(ctx context.Context, req *pb.RecalculateLeaderboardRequest) (*pb.RecalculateLeaderboardResponse, error) {
	// Obtener todos los usuarios ordenados por puntos
	var userStats []models.UserStats
	if err := database.DB.Order("total_points DESC, correct_predictions DESC").Find(&userStats).Error; err != nil {
		log.Printf("Error fetching users for recalculation: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch users: %v", err)
	}

	// Actualizar rangos
	for i := range userStats {
		userStats[i].Rank = i + 1
		if err := database.DB.Model(&userStats[i]).Update("rank", userStats[i].Rank).Error; err != nil {
			log.Printf("Error updating rank for user %s: %v", userStats[i].UserID, err)
		}
	}

	log.Printf("Recalculated leaderboard for %d users", len(userStats))

	return &pb.RecalculateLeaderboardResponse{
		Message:        "Leaderboard recalculated successfully",
		UsersProcessed: int32(len(userStats)),
		GamesEvaluated: 0, // TODO: integrar con game service
	}, nil
}

// ========================================
// Helper Functions
// ========================================

func calculatePercentage(correct, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(correct) / float64(total) * 100.0
}
