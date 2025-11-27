@echo off
echo Testing Game Service gRPC...
echo.

cd /d "%~dp0"

REM Test Game Service gRPC
echo Running comprehensive Game Service gRPC tests...
go run examples/game-grpc-client/main.go

echo.
echo Done!
pause
