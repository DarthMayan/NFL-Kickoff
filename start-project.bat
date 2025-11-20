@echo off
echo ========================================
echo  KICKOFF PROJECT - Docker Compose Start
echo ========================================
echo.

echo Limpiando contenedores anteriores...
docker-compose down --volumes --remove-orphans

echo.
echo Construyendo y ejecutando servicios...
docker-compose up --build -d

echo.
echo Esperando que los servicios esten listos...
timeout /t 10 /nobreak > nul

echo.
echo ========================================
echo  PROYECTO INICIADO EXITOSAMENTE!
echo ========================================
echo.
echo URLs disponibles:
echo - Consul UI: http://localhost:8500
echo - Gateway: http://localhost:8080
echo - User Service: http://localhost:8081/health
echo - Game Service: http://localhost:8082/health
echo - Prediction Service: http://localhost:8083/health
echo - Leaderboard Service: http://localhost:8084/health
echo.
echo Servicios API disponibles via Gateway:
echo - Usuarios: http://localhost:8080/api/users
echo - Equipos: http://localhost:8080/api/teams
echo - Juegos: http://localhost:8080/api/games
echo - Predicciones: http://localhost:8080/api/predictions
echo - Leaderboard: http://localhost:8080/api/leaderboard
echo.
echo Para detener el proyecto, ejecuta: stop-project.bat
echo Para ver logs: docker-compose logs -f [nombre-servicio]
echo.
pause