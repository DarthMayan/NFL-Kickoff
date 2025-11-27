@echo off
echo ========================================
echo Deploying Kickoff NFL to Kubernetes
echo ========================================
echo.

REM Build all Docker images first
echo Step 1: Building Docker images...
cd /d "%~dp0\.."
docker-compose build

echo.
echo Step 2: Creating namespace...
kubectl apply -f k8s/base/namespace.yaml

echo.
echo Step 3: Creating ConfigMaps...
kubectl apply -f k8s/config/configmap.yaml

echo.
echo Step 4: Creating Services...
kubectl apply -f k8s/services/

echo.
echo Step 5: Creating Deployments...
kubectl apply -f k8s/deployments/

echo.
echo Step 6: Creating HorizontalPodAutoscalers...
kubectl apply -f k8s/base/hpa.yaml

echo.
echo ========================================
echo Deployment Complete!
echo ========================================
echo.
echo Checking deployment status...
kubectl get all -n kickoff

echo.
echo To get the Gateway LoadBalancer IP:
echo kubectl get svc gateway-service -n kickoff
echo.
echo To watch pods starting:
echo kubectl get pods -n kickoff -w
echo.
pause
