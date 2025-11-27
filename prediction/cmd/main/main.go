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
	"google.golang.org/protobuf/types/known/timestamppb"

	"kickoff.com/prediction/internal/database"
	"kickoff.com/prediction/internal/models"
	pb "kickoff.com/proto"
)

const serviceName = "prediction"

type PredictionService struct {
	pb.UnimplementedPredictionServiceServer
}

func main() {
	var grpcPort int
	flag.IntVar(&grpcPort, "grpc-port", 9083, "gRPC server port")
	flag.Parse()

	log.Printf("Starting Prediction Service - gRPC:%d", grpcPort)
	log.Printf("Service discovery: Kubernetes DNS")

	// Conectar a la base de datos
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()
	log.Println("✅ Connected to PostgreSQL database")

	predictionService := &PredictionService{}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPredictionServiceServer(grpcServer, predictionService)

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

// ========================================
// gRPC Service Implementation
// ========================================

func (ps *PredictionService) CreatePrediction(ctx context.Context, req *pb.CreatePredictionRequest) (*pb.CreatePredictionResponse, error) {
	if req.UserId == "" || req.GameId == "" || req.PredictedWinnerId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, game_id, and predicted_winner_id are required")
	}

	// Verificar que no exista predicción para este usuario y juego
	var existing models.Prediction
	result := database.DB.Where("user_id = ? AND game_id = ?", req.UserId, req.GameId).First(&existing)
	if result.Error == nil {
		return nil, status.Error(codes.AlreadyExists, "Prediction already exists for this game")
	}

	predictionID := generatePredictionID()
	prediction := models.Prediction{
		ID:                predictionID,
		UserID:            req.UserId,
		GameID:            req.GameId,
		PredictedWinnerID: req.PredictedWinnerId,
		Status:            models.PredictionStatusPending,
		Points:            0,
	}

	if err := database.DB.Create(&prediction).Error; err != nil {
		log.Printf("Error creating prediction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create prediction: %v", err)
	}

	log.Printf("Created prediction: %s for user %s on game %s", prediction.ID, req.UserId, req.GameId)

	return &pb.CreatePredictionResponse{
		Prediction: &pb.Prediction{
			Id:                prediction.ID,
			UserId:            prediction.UserID,
			GameId:            prediction.GameID,
			PredictedWinnerId: prediction.PredictedWinnerID,
			Status:            modelStatusToProto(prediction.Status),
			Points:            int32(prediction.Points),
			CreatedAt:         timestamppb.New(prediction.CreatedAt),
			UpdatedAt:         timestamppb.New(prediction.UpdatedAt),
		},
		Message: "Prediction created successfully",
	}, nil
}

func (ps *PredictionService) GetAllPredictions(ctx context.Context, req *pb.GetAllPredictionsRequest) (*pb.GetAllPredictionsResponse, error) {
	var predictions []models.Prediction
	if err := database.DB.Find(&predictions).Error; err != nil {
		log.Printf("Error fetching predictions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch predictions: %v", err)
	}

	var pbPredictions []*pb.Prediction
	for _, pred := range predictions {
		pbPredictions = append(pbPredictions, &pb.Prediction{
			Id:                pred.ID,
			UserId:            pred.UserID,
			GameId:            pred.GameID,
			PredictedWinnerId: pred.PredictedWinnerID,
			Status:            modelStatusToProto(pred.Status),
			Points:            int32(pred.Points),
			CreatedAt:         timestamppb.New(pred.CreatedAt),
			UpdatedAt:         timestamppb.New(pred.UpdatedAt),
		})
	}

	return &pb.GetAllPredictionsResponse{
		Predictions: pbPredictions,
		Total:       int32(len(pbPredictions)),
	}, nil
}

func (ps *PredictionService) GetPredictionByID(ctx context.Context, req *pb.GetPredictionByIDRequest) (*pb.GetPredictionByIDResponse, error) {
	if req.PredictionId == "" {
		return nil, status.Error(codes.InvalidArgument, "prediction_id is required")
	}

	var prediction models.Prediction
	if err := database.DB.Where("id = ?", req.PredictionId).First(&prediction).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Prediction not found")
	}

	return &pb.GetPredictionByIDResponse{
		Prediction: &pb.Prediction{
			Id:                prediction.ID,
			UserId:            prediction.UserID,
			GameId:            prediction.GameID,
			PredictedWinnerId: prediction.PredictedWinnerID,
			Status:            modelStatusToProto(prediction.Status),
			Points:            int32(prediction.Points),
			CreatedAt:         timestamppb.New(prediction.CreatedAt),
			UpdatedAt:         timestamppb.New(prediction.UpdatedAt),
		},
	}, nil
}

func (ps *PredictionService) GetUserPredictions(ctx context.Context, req *pb.GetUserPredictionsRequest) (*pb.GetUserPredictionsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var predictions []models.Prediction
	if err := database.DB.Where("user_id = ?", req.UserId).Find(&predictions).Error; err != nil {
		log.Printf("Error fetching user predictions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch user predictions: %v", err)
	}

	var pbPredictions []*pb.Prediction
	var correct, incorrect, pending int32
	for _, pred := range predictions {
		pbPredictions = append(pbPredictions, &pb.Prediction{
			Id:                pred.ID,
			UserId:            pred.UserID,
			GameId:            pred.GameID,
			PredictedWinnerId: pred.PredictedWinnerID,
			Status:            modelStatusToProto(pred.Status),
			Points:            int32(pred.Points),
			CreatedAt:         timestamppb.New(pred.CreatedAt),
			UpdatedAt:         timestamppb.New(pred.UpdatedAt),
		})

		switch pred.Status {
		case models.PredictionStatusCorrect:
			correct++
		case models.PredictionStatusIncorrect:
			incorrect++
		case models.PredictionStatusPending:
			pending++
		}
	}

	percentage := 0.0
	if correct+incorrect > 0 {
		percentage = float64(correct) / float64(correct+incorrect) * 100
	}

	return &pb.GetUserPredictionsResponse{
		UserId:      req.UserId,
		Predictions: pbPredictions,
		Total:       int32(len(pbPredictions)),
		Correct:     correct,
		Incorrect:   incorrect,
		Pending:     pending,
		Percentage:  percentage,
	}, nil
}

func (ps *PredictionService) GetGamePredictions(ctx context.Context, req *pb.GetGamePredictionsRequest) (*pb.GetGamePredictionsResponse, error) {
	if req.GameId == "" {
		return nil, status.Error(codes.InvalidArgument, "game_id is required")
	}

	var predictions []models.Prediction
	if err := database.DB.Where("game_id = ?", req.GameId).Find(&predictions).Error; err != nil {
		log.Printf("Error fetching game predictions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch game predictions: %v", err)
	}

	var pbPredictions []*pb.Prediction
	for _, pred := range predictions {
		pbPredictions = append(pbPredictions, &pb.Prediction{
			Id:                pred.ID,
			UserId:            pred.UserID,
			GameId:            pred.GameID,
			PredictedWinnerId: pred.PredictedWinnerID,
			Status:            modelStatusToProto(pred.Status),
			Points:            int32(pred.Points),
			CreatedAt:         timestamppb.New(pred.CreatedAt),
			UpdatedAt:         timestamppb.New(pred.UpdatedAt),
		})
	}

	return &pb.GetGamePredictionsResponse{
		GameId:      req.GameId,
		Predictions: pbPredictions,
		Total:       int32(len(pbPredictions)),
	}, nil
}

func (ps *PredictionService) GetWeekPredictions(ctx context.Context, req *pb.GetWeekPredictionsRequest) (*pb.GetWeekPredictionsResponse, error) {
	// Esta funcionalidad requeriría join con la tabla de games
	// Por ahora retornamos todas las predicciones
	var predictions []models.Prediction
	if err := database.DB.Find(&predictions).Error; err != nil {
		log.Printf("Error fetching week predictions: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch week predictions: %v", err)
	}

	var pbPredictions []*pb.Prediction
	for _, pred := range predictions {
		pbPredictions = append(pbPredictions, &pb.Prediction{
			Id:                pred.ID,
			UserId:            pred.UserID,
			GameId:            pred.GameID,
			PredictedWinnerId: pred.PredictedWinnerID,
			Status:            modelStatusToProto(pred.Status),
			Points:            int32(pred.Points),
			CreatedAt:         timestamppb.New(pred.CreatedAt),
			UpdatedAt:         timestamppb.New(pred.UpdatedAt),
		})
	}

	return &pb.GetWeekPredictionsResponse{
		Week:        req.Week,
		Predictions: pbPredictions,
		Total:       int32(len(pbPredictions)),
	}, nil
}

func (ps *PredictionService) DeletePrediction(ctx context.Context, req *pb.DeletePredictionRequest) (*pb.DeletePredictionResponse, error) {
	if req.PredictionId == "" {
		return nil, status.Error(codes.InvalidArgument, "prediction_id is required")
	}

	var prediction models.Prediction
	if err := database.DB.Where("id = ?", req.PredictionId).First(&prediction).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Prediction not found")
	}

	// Solo permitir eliminar predicciones pendientes
	if prediction.Status != models.PredictionStatusPending {
		return nil, status.Error(codes.FailedPrecondition, "Can only delete pending predictions")
	}

	if err := database.DB.Delete(&prediction).Error; err != nil {
		log.Printf("Error deleting prediction: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete prediction: %v", err)
	}

	log.Printf("Deleted prediction: %s", req.PredictionId)

	return &pb.DeletePredictionResponse{
		Success: true,
		Message: "Prediction deleted successfully",
	}, nil
}

func (ps *PredictionService) UpdatePredictionStatus(ctx context.Context, req *pb.UpdatePredictionStatusRequest) (*pb.UpdatePredictionStatusResponse, error) {
	if req.PredictionId == "" {
		return nil, status.Error(codes.InvalidArgument, "prediction_id is required")
	}

	var prediction models.Prediction
	if err := database.DB.Where("id = ?", req.PredictionId).First(&prediction).Error; err != nil {
		return nil, status.Error(codes.NotFound, "Prediction not found")
	}

	// Actualizar status y puntos
	updates := map[string]interface{}{
		"status": protoStatusToModel(req.Status),
		"points": int(req.Points),
	}

	if err := database.DB.Model(&prediction).Updates(updates).Error; err != nil {
		log.Printf("Error updating prediction status: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update prediction status: %v", err)
	}

	// Recargar
	database.DB.Where("id = ?", req.PredictionId).First(&prediction)

	log.Printf("Updated prediction %s status to %v", req.PredictionId, req.Status)

	return &pb.UpdatePredictionStatusResponse{
		Prediction: &pb.Prediction{
			Id:                prediction.ID,
			UserId:            prediction.UserID,
			GameId:            prediction.GameID,
			PredictedWinnerId: prediction.PredictedWinnerID,
			Status:            modelStatusToProto(prediction.Status),
			Points:            int32(prediction.Points),
			CreatedAt:         timestamppb.New(prediction.CreatedAt),
			UpdatedAt:         timestamppb.New(prediction.UpdatedAt),
		},
		Message: "Prediction status updated successfully",
	}, nil
}

// ========================================
// Helper Functions
// ========================================

func generatePredictionID() string {
	var count int64
	database.DB.Model(&models.Prediction{}).Count(&count)
	return fmt.Sprintf("pred_%d", count+1)
}

func modelStatusToProto(status models.PredictionStatus) pb.PredictionStatus {
	switch status {
	case models.PredictionStatusPending:
		return pb.PredictionStatus_PREDICTION_STATUS_PENDING
	case models.PredictionStatusCorrect:
		return pb.PredictionStatus_PREDICTION_STATUS_CORRECT
	case models.PredictionStatusIncorrect:
		return pb.PredictionStatus_PREDICTION_STATUS_INCORRECT
	case models.PredictionStatusVoid:
		return pb.PredictionStatus_PREDICTION_STATUS_VOID
	default:
		return pb.PredictionStatus_PREDICTION_STATUS_UNSPECIFIED
	}
}

func protoStatusToModel(status pb.PredictionStatus) models.PredictionStatus {
	switch status {
	case pb.PredictionStatus_PREDICTION_STATUS_PENDING:
		return models.PredictionStatusPending
	case pb.PredictionStatus_PREDICTION_STATUS_CORRECT:
		return models.PredictionStatusCorrect
	case pb.PredictionStatus_PREDICTION_STATUS_INCORRECT:
		return models.PredictionStatusIncorrect
	case pb.PredictionStatus_PREDICTION_STATUS_VOID:
		return models.PredictionStatusVoid
	default:
		return models.PredictionStatusPending
	}
}
