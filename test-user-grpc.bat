@echo off
echo Testing User Service gRPC...
echo.

cd /d "%~dp0"

REM Test CreateUser
echo 1. Creating a new user via gRPC...
go run examples/user-grpc-client/main.go

echo.
echo Done!
pause
