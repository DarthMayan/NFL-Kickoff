package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "kickoff.com/proto"
)

func main() {
	// Conectar al servidor gRPC
	conn, err := grpc.NewClient("localhost:9083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Crear cliente
	client := pb.NewPredictionServiceClient(conn)

	// Context con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ========================================
	// TEST 1: Crear Predicción
	// ========================================
	log.Println("========================================")
	log.Println("TEST 1: Creating prediction...")
	log.Println("========================================")

	createReq := &pb.CreatePredictionRequest{
		UserId:            "user123",
		GameId:            "1",
		PredictedWinnerId: "KC",
	}

	createResp, err := client.CreatePrediction(ctx, createReq)
	if err != nil {
		log.Fatalf("CreatePrediction failed: %v", err)
	}

	log.Printf("✅ Prediction created: %v", createResp.Prediction)
	log.Printf("   ID: %s", createResp.Prediction.Id)
	log.Printf("   Message: %s\n", createResp.Message)

	predictionID := createResp.Prediction.Id

	// ========================================
	// TEST 2: Obtener Predicción por ID
	// ========================================
	log.Println("========================================")
	log.Println("TEST 2: Getting prediction by ID...")
	log.Println("========================================")

	getReq := &pb.GetPredictionByIDRequest{
		PredictionId: predictionID,
	}

	getResp, err := client.GetPredictionByID(ctx, getReq)
	if err != nil {
		log.Fatalf("GetPredictionByID failed: %v", err)
	}

	log.Printf("✅ Prediction found: %v\n", getResp.Prediction)

	// ========================================
	// TEST 3: Crear otra predicción para el mismo usuario
	// ========================================
	log.Println("========================================")
	log.Println("TEST 3: Creating another prediction for same user...")
	log.Println("========================================")

	createReq2 := &pb.CreatePredictionRequest{
		UserId:            "user123",
		GameId:            "2",
		PredictedWinnerId: "BUF",
	}

	createResp2, err := client.CreatePrediction(ctx, createReq2)
	if err != nil {
		log.Fatalf("CreatePrediction failed: %v", err)
	}

	log.Printf("✅ Second prediction created: %v\n", createResp2.Prediction)

	// ========================================
	// TEST 4: Obtener todas las predicciones del usuario
	// ========================================
	log.Println("========================================")
	log.Println("TEST 4: Getting user predictions...")
	log.Println("========================================")

	userReq := &pb.GetUserPredictionsRequest{
		UserId: "user123",
	}

	userResp, err := client.GetUserPredictions(ctx, userReq)
	if err != nil {
		log.Fatalf("GetUserPredictions failed: %v", err)
	}

	log.Printf("✅ User predictions:")
	log.Printf("   Total: %d", userResp.Total)
	log.Printf("   Correct: %d", userResp.Correct)
	log.Printf("   Incorrect: %d", userResp.Incorrect)
	log.Printf("   Pending: %d", userResp.Pending)
	log.Printf("   Percentage: %.2f%%", userResp.Percentage)
	for i, pred := range userResp.Predictions {
		log.Printf("   [%d] ID:%s GameID:%s WinnerID:%s Status:%v",
			i+1, pred.Id, pred.GameId, pred.PredictedWinnerId, pred.Status)
	}
	log.Println()

	// ========================================
	// TEST 5: Obtener predicciones de un juego
	// ========================================
	log.Println("========================================")
	log.Println("TEST 5: Getting game predictions...")
	log.Println("========================================")

	gameReq := &pb.GetGamePredictionsRequest{
		GameId: "1",
	}

	gameResp, err := client.GetGamePredictions(ctx, gameReq)
	if err != nil {
		log.Fatalf("GetGamePredictions failed: %v", err)
	}

	log.Printf("✅ Game predictions for game 1:")
	log.Printf("   Total: %d", gameResp.Total)
	for i, pred := range gameResp.Predictions {
		log.Printf("   [%d] UserID:%s PredictedWinner:%s",
			i+1, pred.UserId, pred.PredictedWinnerId)
	}
	log.Println()

	// ========================================
	// TEST 6: Actualizar estado de predicción
	// ========================================
	log.Println("========================================")
	log.Println("TEST 6: Updating prediction status...")
	log.Println("========================================")

	updateReq := &pb.UpdatePredictionStatusRequest{
		PredictionId: predictionID,
		Status:       pb.PredictionStatus_PREDICTION_STATUS_CORRECT,
		Points:       1,
	}

	updateResp, err := client.UpdatePredictionStatus(ctx, updateReq)
	if err != nil {
		log.Fatalf("UpdatePredictionStatus failed: %v", err)
	}

	log.Printf("✅ Prediction status updated: %v", updateResp.Prediction)
	log.Printf("   New Status: %v", updateResp.Prediction.Status)
	log.Printf("   Points: %d\n", updateResp.Prediction.Points)

	// ========================================
	// TEST 7: Listar todas las predicciones
	// ========================================
	log.Println("========================================")
	log.Println("TEST 7: Getting all predictions...")
	log.Println("========================================")

	allReq := &pb.GetAllPredictionsRequest{}

	allResp, err := client.GetAllPredictions(ctx, allReq)
	if err != nil {
		log.Fatalf("GetAllPredictions failed: %v", err)
	}

	log.Printf("✅ All predictions:")
	log.Printf("   Total: %d", allResp.Total)
	for i, pred := range allResp.Predictions {
		log.Printf("   [%d] ID:%s User:%s Game:%s Status:%v",
			i+1, pred.Id, pred.UserId, pred.GameId, pred.Status)
	}
	log.Println()

	// ========================================
	// TEST 8: Intentar crear predicción duplicada (debe fallar)
	// ========================================
	log.Println("========================================")
	log.Println("TEST 8: Trying to create duplicate prediction (should fail)...")
	log.Println("========================================")

	duplicateReq := &pb.CreatePredictionRequest{
		UserId:            "user123",
		GameId:            "1", // Mismo juego que TEST 1
		PredictedWinnerId: "SF",
	}

	_, err = client.CreatePrediction(ctx, duplicateReq)
	if err != nil {
		log.Printf("✅ Expected error received: %v\n", err)
	} else {
		log.Println("❌ ERROR: Duplicate prediction should have failed!")
	}

	// ========================================
	// TEST 9: Eliminar predicción pendiente
	// ========================================
	log.Println("========================================")
	log.Println("TEST 9: Deleting pending prediction...")
	log.Println("========================================")

	// Primero crear una nueva predicción para eliminar
	createReq3 := &pb.CreatePredictionRequest{
		UserId:            "user456",
		GameId:            "1",
		PredictedWinnerId: "KC",
	}

	createResp3, err := client.CreatePrediction(ctx, createReq3)
	if err != nil {
		log.Fatalf("CreatePrediction failed: %v", err)
	}

	deleteID := createResp3.Prediction.Id
	log.Printf("   Created prediction to delete: %s", deleteID)

	deleteReq := &pb.DeletePredictionRequest{
		PredictionId: deleteID,
	}

	deleteResp, err := client.DeletePrediction(ctx, deleteReq)
	if err != nil {
		log.Fatalf("DeletePrediction failed: %v", err)
	}

	log.Printf("✅ Prediction deleted successfully: %s\n", deleteResp.Message)

	// ========================================
	// TEST 10: Intentar eliminar predicción no pendiente (debe fallar)
	// ========================================
	log.Println("========================================")
	log.Println("TEST 10: Trying to delete non-pending prediction (should fail)...")
	log.Println("========================================")

	deleteReq2 := &pb.DeletePredictionRequest{
		PredictionId: predictionID, // Esta tiene status CORRECT del TEST 6
	}

	_, err = client.DeletePrediction(ctx, deleteReq2)
	if err != nil {
		log.Printf("✅ Expected error received: %v\n", err)
	} else {
		log.Println("❌ ERROR: Deleting non-pending prediction should have failed!")
	}

	log.Println("========================================")
	log.Println("✅ ALL TESTS COMPLETED SUCCESSFULLY!")
	log.Println("========================================")
}
