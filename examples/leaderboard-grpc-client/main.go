package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "kickoff.com/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:9084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewLeaderboardServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("========================================")
	fmt.Println("Testing Leaderboard Service gRPC API")
	fmt.Println("========================================")
	fmt.Println()

	// Test 1: Get Full Leaderboard
	fmt.Println("Test 1: GetLeaderboard (no limit)")
	leaderboardResp, err := client.GetLeaderboard(ctx, &pb.GetLeaderboardRequest{})
	if err != nil {
		log.Fatalf("GetLeaderboard failed: %v", err)
	}
	fmt.Printf("✅ Found %d users in leaderboard (%d games finished)\n", leaderboardResp.TotalUsers, leaderboardResp.GamesFinished)
	for i, user := range leaderboardResp.Leaderboard {
		fmt.Printf("   %d. User %s: %d/%d correct (%.1f%%)\n",
			i+1, user.UserId, user.CorrectPicks, user.TotalPicks, user.Percentage)
	}
	fmt.Println()

	// Test 2: Get Top 3 Users
	fmt.Println("Test 2: GetTopUsers (top 3)")
	topUsersResp, err := client.GetTopUsers(ctx, &pb.GetTopUsersRequest{
		TopN: 3,
	})
	if err != nil {
		log.Fatalf("GetTopUsers failed: %v", err)
	}
	fmt.Printf("✅ Top %d users:\n", topUsersResp.Total)
	for i, user := range topUsersResp.TopUsers {
		fmt.Printf("   %d. User %s: %d/%d correct (%.1f%%) - Rank #%d\n",
			i+1, user.UserId, user.CorrectPicks, user.TotalPicks, user.Percentage, user.Rank)
	}
	fmt.Println()

	// Test 3: Get Leaderboard with Pagination (limit 2, offset 0)
	fmt.Println("Test 3: GetLeaderboard (limit 2, offset 0)")
	paginatedResp, err := client.GetLeaderboard(ctx, &pb.GetLeaderboardRequest{
		Limit:  2,
		Offset: 0,
	})
	if err != nil {
		log.Fatalf("GetLeaderboard with pagination failed: %v", err)
	}
	fmt.Printf("✅ Showing %d users (page 1):\n", len(paginatedResp.Leaderboard))
	for _, user := range paginatedResp.Leaderboard {
		fmt.Printf("   Rank #%d: User %s (%.1f%%)\n", user.Rank, user.UserId, user.Percentage)
	}
	fmt.Println()

	// Test 4: Get User Stats (assuming we have a user with ID from leaderboard)
	if len(leaderboardResp.Leaderboard) > 0 {
		testUserID := leaderboardResp.Leaderboard[0].UserId
		fmt.Printf("Test 4: GetUserStats (User: %s)\n", testUserID)
		userStatsResp, err := client.GetUserStats(ctx, &pb.GetUserStatsRequest{
			UserId: testUserID,
		})
		if err != nil {
			log.Fatalf("GetUserStats failed: %v", err)
		}
		fmt.Printf("✅ User Stats:\n")
		fmt.Printf("   User ID: %s\n", userStatsResp.UserStats.UserId)
		fmt.Printf("   Correct Picks: %d/%d (%.1f%%)\n",
			userStatsResp.UserStats.CorrectPicks,
			userStatsResp.UserStats.TotalPicks,
			userStatsResp.UserStats.Percentage)
		fmt.Printf("   Rank: #%d\n", userStatsResp.UserStats.Rank)
		fmt.Printf("   Total Predictions: %d\n", userStatsResp.TotalPredictions)
		fmt.Println("   Predictions:")
		for i, pred := range userStatsResp.Predictions {
			if pred.GameStatus == "finished" {
				correctMark := "✗"
				if pred.Correct {
					correctMark = "✓"
				}
				fmt.Printf("     %d. Game %s: Predicted %s, Actual %s %s\n",
					i+1, pred.GameId, pred.PredictedWinner, pred.ActualWinner, correctMark)
			} else {
				fmt.Printf("     %d. Game %s: Predicted %s (pending)\n",
					i+1, pred.GameId, pred.PredictedWinner)
			}
		}
		fmt.Println()

		// Test 5: Get User Rank
		fmt.Printf("Test 5: GetUserRank (User: %s)\n", testUserID)
		userRankResp, err := client.GetUserRank(ctx, &pb.GetUserRankRequest{
			UserId: testUserID,
		})
		if err != nil {
			log.Fatalf("GetUserRank failed: %v", err)
		}
		fmt.Printf("✅ User Rank: #%d out of %d users\n", userRankResp.Rank, userRankResp.TotalUsers)
		fmt.Printf("   Performance: %d/%d correct (%.1f%%)\n",
			userRankResp.UserScore.CorrectPicks,
			userRankResp.UserScore.TotalPicks,
			userRankResp.UserScore.Percentage)
		fmt.Println()
	}

	// Test 6: Recalculate Leaderboard
	fmt.Println("Test 6: RecalculateLeaderboard")
	recalcResp, err := client.RecalculateLeaderboard(ctx, &pb.RecalculateLeaderboardRequest{})
	if err != nil {
		log.Fatalf("RecalculateLeaderboard failed: %v", err)
	}
	fmt.Printf("✅ %s\n", recalcResp.Message)
	fmt.Printf("   Users Processed: %d\n", recalcResp.UsersProcessed)
	fmt.Printf("   Games Evaluated: %d\n", recalcResp.GamesEvaluated)
	fmt.Println()

	// Test 7: Get Top 10 Users
	fmt.Println("Test 7: GetTopUsers (top 10)")
	top10Resp, err := client.GetTopUsers(ctx, &pb.GetTopUsersRequest{
		TopN: 10,
	})
	if err != nil {
		log.Fatalf("GetTopUsers failed: %v", err)
	}
	fmt.Printf("✅ Top %d users:\n", top10Resp.Total)
	for i, user := range top10Resp.TopUsers {
		fmt.Printf("   %d. User %s: %.1f%% (%d/%d)\n",
			i+1, user.UserId, user.Percentage, user.CorrectPicks, user.TotalPicks)
	}
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("✅ ALL TESTS PASSED!")
	fmt.Println("========================================")
}
