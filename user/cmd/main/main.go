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

	"kickoff.com/user/internal/database"
	"kickoff.com/user/internal/models"
	pb "kickoff.com/proto"
)

const serviceName = "user"

type UserService struct {
	pb.UnimplementedUserServiceServer
}

func main() {
	var grpcPort int
	flag.IntVar(&grpcPort, "grpc-port", 9081, "gRPC server port")
	flag.Parse()

	log.Printf("Starting User Service - gRPC:%d", grpcPort)
	log.Printf("Service discovery: Kubernetes DNS")

	// Conectar a la base de datos
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL database")

	// Inicializar servicio
	userService := &UserService{}

	// Crear listener para gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Crear servidor gRPC
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userService)

	// Registrar health check
	healthServer := health.NewServer()
	healthServer.SetServingStatus(serviceName, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("✅ gRPC server listening on :%d", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for termination signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	database.Close()
	grpcServer.GracefulStop()
}

// ========================================
// gRPC Service Implementation
// ========================================

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	// Verificar que username sea único
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists: %s", req.Username)
	}

	// Verificar que email sea único
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "email already exists: %s", req.Email)
	}

	// Crear nuevo usuario
	user := models.User{
		ID:       generateUserID(),
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		Active:   true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	log.Printf("Created user: %s (%s)", user.Username, user.ID)

	return &pb.CreateUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		},
		Message: "User created successfully",
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, req *pb.GetUserByIDRequest) (*pb.GetUserByIDResponse, error) {
	var user models.User
	if err := database.DB.Where("id = ?", req.UserId).First(&user).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %s", req.UserId)
	}

	return &pb.GetUserByIDResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		},
	}, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, req *pb.GetAllUsersRequest) (*pb.GetAllUsersResponse, error) {
	var users []models.User
	query := database.DB

	if req.ActiveOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Find(&users).Error; err != nil {
		log.Printf("Error fetching users: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to fetch users: %v", err)
	}

	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		})
	}

	return &pb.GetAllUsersResponse{
		Users: pbUsers,
		Total: int32(len(pbUsers)),
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	var user models.User
	if err := database.DB.Where("id = ?", req.UserId).First(&user).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %s", req.UserId)
	}

	// Actualizar campos
	updates := make(map[string]interface{})
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		log.Printf("Error updating user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Recargar usuario actualizado
	database.DB.Where("id = ?", req.UserId).First(&user)

	log.Printf("Updated user: %s", req.UserId)

	return &pb.UpdateUserResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		},
		Message: "User updated successfully",
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	var user models.User
	if err := database.DB.Where("id = ?", req.UserId).First(&user).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %s", req.UserId)
	}

	// Soft delete - marcar como inactivo
	if err := database.DB.Model(&user).Update("active", false).Error; err != nil {
		log.Printf("Error deleting user: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	log.Printf("Deleted (soft) user: %s", req.UserId)

	return &pb.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (s *UserService) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	var users []models.User
	searchTerm := "%" + req.SearchTerm + "%"

	if err := database.DB.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?",
		searchTerm, searchTerm, searchTerm).Find(&users).Error; err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to search users: %v", err)
	}

	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		})
	}

	return &pb.SearchUsersResponse{
		Users:      pbUsers,
		Total:      int32(len(pbUsers)),
		SearchTerm: req.SearchTerm,
	}, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserByUsernameResponse, error) {
	var user models.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found with username: %s", req.Username)
	}

	return &pb.GetUserByUsernameResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		},
	}, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.GetUserByEmailResponse, error) {
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found with email: %s", req.Email)
	}

	return &pb.GetUserByEmailResponse{
		User: &pb.User{
			Id:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			CreatedAt: timestamppb.New(user.CreatedAt),
			Active:    user.Active,
		},
	}, nil
}

// ========================================
// Helper Functions
// ========================================

func generateUserID() string {
	// Generar ID único basado en timestamp
	var count int64
	database.DB.Model(&models.User{}).Count(&count)
	return fmt.Sprintf("user_%d", count+1)
}
