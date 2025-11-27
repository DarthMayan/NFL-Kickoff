@echo off
echo ========================================
echo K6 Load Test - Kickoff NFL
echo ========================================
echo.

REM Verificar que k6 está instalado
set K6_PATH="C:\Program Files\k6\k6.exe"
if not exist %K6_PATH% (
    echo ERROR: K6 no está instalado!
    echo.
    echo Instalalo con:
    echo   winget install k6
    echo.
    echo O descargalo de: https://k6.io/docs/get-started/installation/
    pause
    exit /b 1
)

echo K6 instalado correctamente
echo.

REM Menú de opciones
echo Selecciona el tipo de test:
echo.
echo [1] Load Test Básico (20 usuarios, 1 minuto)
echo [2] Stress Test Completo (0-100 usuarios, ~13 minutos)
echo [3] Test Rápido (10 usuarios, 30 segundos)
echo [4] Test Personalizado
echo.
set /p choice="Ingresa tu opción (1-4): "

if "%choice%"=="1" goto basic
if "%choice%"=="2" goto stress
if "%choice%"=="3" goto quick
if "%choice%"=="4" goto custom
echo Opción inválida
pause
exit /b 1

:basic
echo.
echo Ejecutando Load Test Básico...
echo.
%K6_PATH% run k6-load-test.js
goto end

:stress
echo.
echo ========================================
echo IMPORTANTE: Este test durará ~13 minutos
echo ========================================
echo.
echo Asegúrate de tener:
echo  1. Port-forward activo: kubectl port-forward -n kickoff svc/gateway-service 8080:8080
echo  2. Terminal monitoreando HPA: kubectl get hpa -n kickoff -w
echo  3. Terminal monitoreando pods: kubectl get pods -n kickoff -w
echo.
set /p confirm="¿Continuar? (S/N): "
if /i not "%confirm%"=="S" (
    echo Test cancelado
    pause
    exit /b 0
)
echo.
echo Ejecutando Stress Test...
echo.
%K6_PATH% run k6-stress-test.js
goto end

:quick
echo.
echo Ejecutando Test Rápido...
echo.
%K6_PATH% run --vus 10 --duration 30s k6-load-test.js
goto end

:custom
echo.
echo Test Personalizado
echo.
set /p vus="Número de usuarios virtuales: "
set /p duration="Duración (ej: 30s, 1m, 2m): "
echo.
echo Ejecutando con %vus% usuarios por %duration%...
echo.
%K6_PATH% run --vus %vus% --duration %duration% k6-load-test.js
goto end

:end
echo.
echo ========================================
echo Test Completado
echo ========================================
echo.
echo Para ver el estado del cluster:
echo   kubectl get hpa -n kickoff
echo   kubectl get pods -n kickoff
echo   kubectl top pods -n kickoff
echo.
pause
