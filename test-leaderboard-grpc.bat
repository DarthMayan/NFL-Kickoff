@echo off
echo Testing Leaderboard Service gRPC...
echo.

cd /d "%~dp0"

REM Test Leaderboard Service gRPC
echo Running comprehensive Leaderboard Service gRPC tests...
go run examples/leaderboard-grpc-client/main.go

echo.
echo Done!
pause
