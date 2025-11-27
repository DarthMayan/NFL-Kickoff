@echo off
echo ========================================
echo Kickoff NFL - Aggressive Load Testing
echo ========================================
echo.
echo This script will spawn multiple parallel
echo processes to generate heavy load and
echo trigger HPA autoscaling.
echo.

REM Get the NodePort for gateway service
echo Getting Gateway Service NodePort...
for /f "tokens=5 delims=:/" %%a in ('kubectl get svc gateway-service -n kickoff -o jsonpath^="{.spec.ports[0].nodePort}"') do set NODEPORT=%%a
echo Gateway NodePort: %NODEPORT%
echo.

REM Set the target URL
set TARGET_URL=http://localhost:%NODEPORT%

echo Target URL: %TARGET_URL%
echo.
echo Starting 10 parallel load generators...
echo Press Ctrl+C to stop all processes
echo.
echo ========================================

REM Start 10 background processes
for /L %%i in (1,1,10) do (
    start /B cmd /c "for /L %%j in (1,1,10000) do curl -s -o nul %TARGET_URL%/teams"
)

REM Wait and monitor HPA
echo.
echo Load test running...
echo Monitoring HPA status (refresh every 5 seconds)
echo.

:MONITOR
kubectl get hpa -n kickoff
echo.
echo Pods status:
kubectl get pods -n kickoff
echo.
echo ----------------------------------------
timeout /t 5 /nobreak >nul
goto MONITOR
