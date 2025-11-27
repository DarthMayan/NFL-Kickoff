@echo off
REM Script para generar código Go desde archivos .proto
REM Windows batch script

echo ========================================
echo Generando codigo Go desde archivos .proto
echo ========================================
echo.

REM Verificar que protoc está instalado
where protoc >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: protoc no esta instalado
    echo Por favor instala Protocol Buffers Compiler:
    echo https://github.com/protocolbuffers/protobuf/releases
    echo.
    echo O usa chocolatey: choco install protoc
    exit /b 1
)

echo [OK] protoc encontrado:
protoc --version
echo.

REM Verificar que los plugins de Go están instalados
where protoc-gen-go >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] protoc-gen-go no encontrado, instalando...
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)

where protoc-gen-go-grpc >nul 2>nul
if %errorlevel% neq 0 (
    echo [WARN] protoc-gen-go-grpc no encontrado, instalando...
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
)

echo [OK] Plugins de Go instalados
echo.

REM Generar código
echo Generando codigo para prediction_service.proto...
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/prediction_service.proto

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Fallo la generacion de codigo
    exit /b 1
)

echo.
echo ========================================
echo Codigo generado exitosamente!
echo ========================================
echo.
echo Archivos creados:
dir /b proto\*.pb.go
echo.
echo Ejecuta 'go mod tidy' para actualizar dependencias
