# Consul
run-consul-dev-server:
	docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul consul:1.15.4 agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0

stop-consul:
	docker stop dev-consul && docker rm dev-consul

# Services
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