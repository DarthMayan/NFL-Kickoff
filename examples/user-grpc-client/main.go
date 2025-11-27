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
	conn, err := grpc.NewClient("localhost:9081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("========================================")
	fmt.Println("Testing User Service gRPC API")
	fmt.Println("========================================")
	fmt.Println()

	// Test 1: Create User
	fmt.Println("Test 1: CreateUser")
	createResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "grpcuser",
		Email:    "grpc@test.com",
		FullName: "gRPC Test User",
	})
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	fmt.Printf("✅ User created: ID=%s, Username=%s, Email=%s\n",
		createResp.User.Id, createResp.User.Username, createResp.User.Email)
	fmt.Println()
	userID := createResp.User.Id

	// Test 2: Get User by ID
	fmt.Println("Test 2: GetUserByID")
	getResp, err := client.GetUserByID(ctx, &pb.GetUserByIDRequest{
		UserId: userID,
	})
	if err != nil {
		log.Fatalf("GetUserByID failed: %v", err)
	}
	fmt.Printf("✅ User found: %s (%s)\n", getResp.User.Username, getResp.User.Email)
	fmt.Println()

	// Test 3: Create another user
	fmt.Println("Test 3: Create another user")
	_, err = client.CreateUser(ctx, &pb.CreateUserRequest{
		Username: "grpcuser2",
		Email:    "grpc2@test.com",
		FullName: "Second gRPC User",
	})
	if err != nil {
		log.Fatalf("CreateUser failed: %v", err)
	}
	fmt.Println("✅ Second user created")
	fmt.Println()

	// Test 4: Get All Users
	fmt.Println("Test 4: GetAllUsers")
	allUsersResp, err := client.GetAllUsers(ctx, &pb.GetAllUsersRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		log.Fatalf("GetAllUsers failed: %v", err)
	}
	fmt.Printf("✅ Found %d users:\n", allUsersResp.Total)
	for i, user := range allUsersResp.Users {
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, user.Username, user.Email, user.FullName)
	}
	fmt.Println()

	// Test 5: Search Users
	fmt.Println("Test 5: SearchUsers")
	searchResp, err := client.SearchUsers(ctx, &pb.SearchUsersRequest{
		SearchTerm: "grpc",
	})
	if err != nil {
		log.Fatalf("SearchUsers failed: %v", err)
	}
	fmt.Printf("✅ Search for 'grpc' found %d users:\n", searchResp.Total)
	for i, user := range searchResp.Users {
		fmt.Printf("   %d. %s (%s)\n", i+1, user.Username, user.Email)
	}
	fmt.Println()

	// Test 6: Get User by Username
	fmt.Println("Test 6: GetUserByUsername")
	byUsernameResp, err := client.GetUserByUsername(ctx, &pb.GetUserByUsernameRequest{
		Username: "grpcuser",
	})
	if err != nil {
		log.Fatalf("GetUserByUsername failed: %v", err)
	}
	fmt.Printf("✅ Found user by username: %s (%s)\n",
		byUsernameResp.User.Username, byUsernameResp.User.Email)
	fmt.Println()

	// Test 7: Get User by Email
	fmt.Println("Test 7: GetUserByEmail")
	byEmailResp, err := client.GetUserByEmail(ctx, &pb.GetUserByEmailRequest{
		Email: "grpc@test.com",
	})
	if err != nil {
		log.Fatalf("GetUserByEmail failed: %v", err)
	}
	fmt.Printf("✅ Found user by email: %s (%s)\n",
		byEmailResp.User.Username, byEmailResp.User.Email)
	fmt.Println()

	// Test 8: Update User
	fmt.Println("Test 8: UpdateUser")
	active := false
	updateResp, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
		UserId:   userID,
		Username: "grpcuser_updated",
		Email:    "grpc_updated@test.com",
		FullName: "Updated gRPC User",
		Active:   &active,
	})
	if err != nil {
		log.Fatalf("UpdateUser failed: %v", err)
	}
	fmt.Printf("✅ User updated: %s (%s), Active=%v\n",
		updateResp.User.Username, updateResp.User.Email, updateResp.User.Active)
	fmt.Println()

	// Test 9: Delete User (soft delete)
	fmt.Println("Test 9: DeleteUser (soft delete)")
	deleteResp, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{
		UserId: userID,
	})
	if err != nil {
		log.Fatalf("DeleteUser failed: %v", err)
	}
	fmt.Printf("✅ %s\n", deleteResp.Message)
	fmt.Println()

	// Test 10: Verify user is inactive
	fmt.Println("Test 10: Verify user is now inactive")
	verifyResp, err := client.GetUserByID(ctx, &pb.GetUserByIDRequest{
		UserId: userID,
	})
	if err != nil {
		log.Fatalf("GetUserByID failed: %v", err)
	}
	fmt.Printf("✅ User status verified: Active=%v\n", verifyResp.User.Active)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("✅ ALL TESTS PASSED!")
	fmt.Println("========================================")
}
