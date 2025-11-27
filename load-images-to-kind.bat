@echo off
echo Loading Docker images to Kind cluster...
echo.

echo [1/6] Loading Gateway Service...
kind load docker-image kickoff-gateway-service:latest --name kickoff

echo [2/6] Loading User Service...
kind load docker-image kickoff-user-service:latest --name kickoff

echo [3/6] Loading Game Service...
kind load docker-image kickoff-game-service:latest --name kickoff

echo [4/6] Loading Prediction Service...
kind load docker-image kickoff-prediction-service:latest --name kickoff

echo [5/6] Loading Leaderboard Service...
kind load docker-image kickoff-leaderboard-service:latest --name kickoff

echo [6/6] Loading PostgreSQL...
kind load docker-image postgres:15-alpine --name kickoff

echo.
echo ✅ All images loaded successfully!
pause
