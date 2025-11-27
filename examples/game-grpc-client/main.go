package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "kickoff.com/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9082", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewGameServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("========================================")
	fmt.Println("Testing Game Service gRPC API")
	fmt.Println("========================================")
	fmt.Println()

	// Test 1: Get All Teams
	fmt.Println("Test 1: GetAllTeams")
	teamsResp, err := client.GetAllTeams(ctx, &pb.GetAllTeamsRequest{})
	if err != nil {
		log.Fatalf("GetAllTeams failed: %v", err)
	}
	fmt.Printf("✅ Found %d teams\n", teamsResp.Total)
	if len(teamsResp.Teams) > 0 {
		fmt.Printf("   Sample: %s - %s, %s (%s)\n",
			teamsResp.Teams[0].Id,
			teamsResp.Teams[0].Name,
			teamsResp.Teams[0].City,
			teamsResp.Teams[0].Conference)
	}
	fmt.Println()

	// Test 2: Get Team by ID
	fmt.Println("Test 2: GetTeamByID (KC)")
	teamResp, err := client.GetTeamByID(ctx, &pb.GetTeamByIDRequest{
		TeamId: "KC",
	})
	if err != nil {
		log.Fatalf("GetTeamByID failed: %v", err)
	}
	fmt.Printf("✅ Team found: %s - %s\n", teamResp.Team.Id, teamResp.Team.Name)
	fmt.Printf("   Stadium: %s\n", teamResp.Team.Stadium)
	fmt.Println()

	// Test 3: Get Teams by Conference
	fmt.Println("Test 3: GetTeamsByConference (AFC)")
	afcTeamsResp, err := client.GetTeamsByConference(ctx, &pb.GetTeamsByConferenceRequest{
		Conference: pb.Conference_CONFERENCE_AFC,
	})
	if err != nil {
		log.Fatalf("GetTeamsByConference failed: %v", err)
	}
	fmt.Printf("✅ Found %d AFC teams\n", afcTeamsResp.Total)
	fmt.Println()

	// Test 4: Get Teams by Division
	fmt.Println("Test 4: GetTeamsByDivision (AFC West)")
	afcWestResp, err := client.GetTeamsByDivision(ctx, &pb.GetTeamsByDivisionRequest{
		Division: pb.Division_DIVISION_AFC_WEST,
	})
	if err != nil {
		log.Fatalf("GetTeamsByDivision failed: %v", err)
	}
	fmt.Printf("✅ Found %d teams in AFC West:\n", afcWestResp.Total)
	for i, team := range afcWestResp.Teams {
		fmt.Printf("   %d. %s - %s\n", i+1, team.Id, team.Name)
	}
	fmt.Println()

	// Test 5: Get All Games
	fmt.Println("Test 5: GetAllGames")
	gamesResp, err := client.GetAllGames(ctx, &pb.GetAllGamesRequest{})
	if err != nil {
		log.Fatalf("GetAllGames failed: %v", err)
	}
	fmt.Printf("✅ Found %d games\n", gamesResp.Total)
	for i, game := range gamesResp.Games {
		fmt.Printf("   %d. %s: %s vs %s (Week %d) - Status: %s, Score: %d-%d\n",
			i+1, game.Id, game.HomeTeamId, game.AwayTeamId,
			game.Week, game.Status, game.HomeScore, game.AwayScore)
	}
	fmt.Println()

	// Test 6: Get Game by ID
	fmt.Println("Test 6: GetGameByID (game_1)")
	gameResp, err := client.GetGameByID(ctx, &pb.GetGameByIDRequest{
		GameId: "game_1",
	})
	if err != nil {
		log.Fatalf("GetGameByID failed: %v", err)
	}
	fmt.Printf("✅ Game found: %s vs %s (Week %d)\n",
		gameResp.Game.HomeTeamId, gameResp.Game.AwayTeamId, gameResp.Game.Week)
	fmt.Println()

	// Test 7: Get Games by Week
	fmt.Println("Test 7: GetGamesByWeek (Week 1)")
	weekGamesResp, err := client.GetGamesByWeek(ctx, &pb.GetGamesByWeekRequest{
		Week: 1,
	})
	if err != nil {
		log.Fatalf("GetGamesByWeek failed: %v", err)
	}
	fmt.Printf("✅ Found %d games in Week 1\n", weekGamesResp.Total)
	fmt.Println()

	// Test 8: Get Games by Team
	fmt.Println("Test 8: GetGamesByTeam (KC)")
	teamGamesResp, err := client.GetGamesByTeam(ctx, &pb.GetGamesByTeamRequest{
		TeamId: "KC",
	})
	if err != nil {
		log.Fatalf("GetGamesByTeam failed: %v", err)
	}
	fmt.Printf("✅ Found %d games for KC\n", teamGamesResp.Total)
	fmt.Println()

	// Test 9: Get Games by Status
	fmt.Println("Test 9: GetGamesByStatus (SCHEDULED)")
	statusGamesResp, err := client.GetGamesByStatus(ctx, &pb.GetGamesByStatusRequest{
		Status: pb.GameStatus_GAME_STATUS_SCHEDULED,
	})
	if err != nil {
		log.Fatalf("GetGamesByStatus failed: %v", err)
	}
	fmt.Printf("✅ Found %d scheduled games\n", statusGamesResp.Total)
	fmt.Println()

	// Test 10: Create Game
	fmt.Println("Test 10: CreateGame")
	scheduledTime := time.Now().Add(7 * 24 * time.Hour) // 1 week from now
	createGameResp, err := client.CreateGame(ctx, &pb.CreateGameRequest{
		HomeTeamId:  "BUF",
		AwayTeamId:  "MIA",
		Week:        2,
		ScheduledAt: timestamppb.New(scheduledTime),
	})
	if err != nil {
		log.Fatalf("CreateGame failed: %v", err)
	}
	fmt.Printf("✅ %s\n", createGameResp.Message)
	fmt.Printf("   Game ID: %s, %s vs %s (Week %d)\n",
		createGameResp.Game.Id,
		createGameResp.Game.HomeTeamId,
		createGameResp.Game.AwayTeamId,
		createGameResp.Game.Week)
	newGameID := createGameResp.Game.Id
	fmt.Println()

	// Test 11: Update Game Score
	fmt.Println("Test 11: UpdateGameScore")
	updateScoreResp, err := client.UpdateGameScore(ctx, &pb.UpdateGameScoreRequest{
		GameId:    newGameID,
		HomeScore: 24,
		AwayScore: 17,
	})
	if err != nil {
		log.Fatalf("UpdateGameScore failed: %v", err)
	}
	fmt.Printf("✅ %s\n", updateScoreResp.Message)
	fmt.Printf("   Score: %s %d - %d %s\n",
		updateScoreResp.Game.HomeTeamId,
		updateScoreResp.Game.HomeScore,
		updateScoreResp.Game.AwayScore,
		updateScoreResp.Game.AwayTeamId)
	fmt.Println()

	// Test 12: Update Game Status to IN_PROGRESS
	fmt.Println("Test 12: UpdateGameStatus (IN_PROGRESS)")
	updateStatusResp, err := client.UpdateGameStatus(ctx, &pb.UpdateGameStatusRequest{
		GameId: newGameID,
		Status: pb.GameStatus_GAME_STATUS_IN_PROGRESS,
	})
	if err != nil {
		log.Fatalf("UpdateGameStatus failed: %v", err)
	}
	fmt.Printf("✅ %s\n", updateStatusResp.Message)
	fmt.Printf("   Game status: %s\n", updateStatusResp.Game.Status)
	fmt.Println()

	// Test 13: Update Game Status to COMPLETED
	fmt.Println("Test 13: UpdateGameStatus (COMPLETED)")
	finalStatusResp, err := client.UpdateGameStatus(ctx, &pb.UpdateGameStatusRequest{
		GameId: newGameID,
		Status: pb.GameStatus_GAME_STATUS_COMPLETED,
	})
	if err != nil {
		log.Fatalf("UpdateGameStatus failed: %v", err)
	}
	fmt.Printf("✅ %s\n", finalStatusResp.Message)
	fmt.Printf("   Game status: %s\n", finalStatusResp.Game.Status)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("✅ ALL TESTS PASSED!")
	fmt.Println("========================================")
}
