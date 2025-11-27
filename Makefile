# ========================================
# PROTO / gRPC CODE GENERATION
# ========================================

# Install protoc dependencies (run once)
proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
proto-gen:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/prediction_service.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/user_service.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/game_service.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/leaderboard_service.proto

# Clean generated proto files
proto-clean:
	rm -f proto/*.pb.go

# Generate all proto files (add more as needed)
proto-gen-all: proto-gen

# ========================================
# KIND CLUSTER MANAGEMENT
# ========================================

kind-create:
	kind create cluster --name kickoff --config kind-config.yaml

kind-delete:
	kind delete cluster --name kickoff

kind-load-images:
	kind load docker-image kickoff-gateway-service:latest --name kickoff
	kind load docker-image kickoff-user-service:latest --name kickoff
	kind load docker-image kickoff-game-service:latest --name kickoff
	kind load docker-image kickoff-prediction-service:latest --name kickoff
	kind load docker-image kickoff-leaderboard-service:latest --name kickoff

# ========================================
# SERVICES (Development Mode)
# ========================================
user-service:
	go run user/cmd/main/main.go

game-service:
	go run game/cmd/main/main.go

prediction-service:
	go run prediction/cmd/main/main.go

leaderboard-service:
	go run leaderboard/cmd/main/main.go

gateway-service:
	go run gateway/cmd/main/main.go