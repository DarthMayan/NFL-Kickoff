@echo off
REM Script para probar el servidor gRPC del Prediction Service
REM Requiere que grpcurl esté instalado: choco install grpcurl

echo ========================================
echo Testing Prediction Service gRPC API
echo ========================================
echo.

REM Verificar que grpcurl está instalado
where grpcurl >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: grpcurl no esta instalado
    echo Por favor instala grpcurl:
    echo   choco install grpcurl
    echo.
    echo O descarga desde: https://github.com/fullstorydev/grpcurl/releases
    exit /b 1
)

echo [OK] grpcurl encontrado
echo.

REM Verificar que el servidor está corriendo
echo Verificando que el servidor gRPC esta corriendo en localhost:9083...
grpcurl -plaintext localhost:9083 list >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: No se puede conectar al servidor gRPC en localhost:9083
    echo.
    echo Asegurate de que el Prediction Service este corriendo:
    echo   go run prediction/cmd/main/main.go
    echo.
    echo O con Docker:
    echo   docker-compose up prediction-service
    exit /b 1
)

echo [OK] Servidor gRPC esta corriendo
echo.

REM Listar servicios disponibles
echo ========================================
echo 1. Listando servicios disponibles...
echo ========================================
grpcurl -plaintext localhost:9083 list
echo.

REM Listar métodos del servicio
echo ========================================
echo 2. Listando metodos de PredictionService...
echo ========================================
grpcurl -plaintext localhost:9083 list prediction.PredictionService
echo.

REM Crear una predicción
echo ========================================
echo 3. Creando prediccion para user1...
echo ========================================
grpcurl -plaintext -d "{\"user_id\":\"user1\", \"game_id\":\"1\", \"predicted_winner_id\":\"KC\"}" localhost:9083 prediction.PredictionService/CreatePrediction
echo.

REM Crear otra predicción
echo ========================================
echo 4. Creando prediccion para user1 en juego 2...
echo ========================================
grpcurl -plaintext -d "{\"user_id\":\"user1\", \"game_id\":\"2\", \"predicted_winner_id\":\"BUF\"}" localhost:9083 prediction.PredictionService/CreatePrediction
echo.

REM Obtener predicciones del usuario
echo ========================================
echo 5. Obteniendo predicciones de user1...
echo ========================================
grpcurl -plaintext -d "{\"user_id\":\"user1\"}" localhost:9083 prediction.PredictionService/GetUserPredictions
echo.

REM Obtener todas las predicciones
echo ========================================
echo 6. Obteniendo todas las predicciones...
echo ========================================
grpcurl -plaintext -d "{}" localhost:9083 prediction.PredictionService/GetAllPredictions
echo.

echo ========================================
echo Pruebas completadas!
echo ========================================
echo.
echo Para mas pruebas, puedes usar:
echo   grpcurl -plaintext -d "..." localhost:9083 prediction.PredictionService/[METODO]
echo.
echo Metodos disponibles:
echo   - CreatePrediction
echo   - GetPredictionByID
echo   - GetUserPredictions
echo   - GetGamePredictions
echo   - GetWeekPredictions
echo   - GetAllPredictions
echo   - DeletePrediction
echo   - UpdatePredictionStatus
