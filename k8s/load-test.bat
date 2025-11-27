@echo off
echo ========================================
echo Kickoff NFL - Load Testing Script
echo ========================================
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
echo Press Ctrl+C to stop the load test
echo.
echo ========================================
echo Starting Load Test...
echo ========================================
echo.

REM Counter for requests
set /a COUNT=0
set /a ERRORS=0

:LOOP
set /a COUNT=%COUNT%+1

REM Test different endpoints in rotation
set /a ENDPOINT_NUM=%COUNT% %% 5

if %ENDPOINT_NUM%==0 (
    curl -s -o nul -w "Request %COUNT%: GET /teams - Status: %%{http_code} - Time: %%{time_total}s\n" %TARGET_URL%/teams
) else if %ENDPOINT_NUM%==1 (
    curl -s -o nul -w "Request %COUNT%: GET /users - Status: %%{http_code} - Time: %%{time_total}s\n" %TARGET_URL%/users
) else if %ENDPOINT_NUM%==2 (
    curl -s -o nul -w "Request %COUNT%: GET /games - Status: %%{http_code} - Time: %%{time_total}s\n" %TARGET_URL%/games
) else if %ENDPOINT_NUM%==3 (
    curl -s -o nul -w "Request %COUNT%: GET /leaderboard - Status: %%{http_code} - Time: %%{time_total}s\n" %TARGET_URL%/leaderboard
) else (
    curl -s -o nul -w "Request %COUNT%: GET /teams/1 - Status: %%{http_code} - Time: %%{time_total}s\n" %TARGET_URL%/teams/1
)

REM Small delay to not overwhelm (adjust or remove for more aggressive testing)
timeout /t 0 /nobreak >nul

REM Continue loop
goto LOOP
