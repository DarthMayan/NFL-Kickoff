@echo off
echo ========================================
echo  KICKOFF PROJECT - Docker Compose Stop
echo ========================================
echo.

echo Deteniendo y limpiando todos los servicios...
docker-compose down --volumes --remove-orphans

echo.
echo Limpiando imágenes no utilizadas (opcional)...
docker image prune -f

echo.
echo Proyecto detenido y limpiado.
echo.
pause