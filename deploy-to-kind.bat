@echo off
echo Deploying to Kind cluster...
echo.

echo [Step 1] Creating namespace...
kubectl create namespace kickoff --dry-run=client -o yaml | kubectl apply -f -

echo [Step 2] Applying ConfigMaps...
kubectl apply -f k8s/config/configmap.yaml
kubectl apply -f k8s/config/postgres-config.yaml

echo [Step 3] Applying PersistentVolumeClaims...
kubectl apply -f k8s/base/postgres-pvc.yaml

echo [Step 4] Deploying PostgreSQL...
kubectl apply -f k8s/deployments/postgres-deployment.yaml
kubectl apply -f k8s/services/postgres-service.yaml

echo [Step 5] Waiting for PostgreSQL to be ready...
kubectl wait --for=condition=ready pod -l app=postgres -n kickoff --timeout=120s

echo [Step 6] Initializing PostgreSQL schema...
type db\init-schema.sql | kubectl exec -i -n kickoff deployment/postgres -- psql -U kickoff_user -d kickoff_nfl

echo [Step 7] Deploying microservices...
kubectl apply -f k8s/deployments/user-deployment.yaml
kubectl apply -f k8s/deployments/game-deployment.yaml
kubectl apply -f k8s/deployments/prediction-deployment.yaml
kubectl apply -f k8s/deployments/leaderboard-deployment.yaml
kubectl apply -f k8s/deployments/gateway-deployment.yaml

echo [Step 8] Deploying services...
kubectl apply -f k8s/services/user-service.yaml
kubectl apply -f k8s/services/game-service.yaml
kubectl apply -f k8s/services/prediction-service.yaml
kubectl apply -f k8s/services/leaderboard-service.yaml
kubectl apply -f k8s/services/gateway-service.yaml

echo [Step 9] Deploying HPAs...
kubectl apply -f k8s/hpa/gateway-hpa.yaml
kubectl apply -f k8s/hpa/prediction-hpa.yaml

echo.
echo ✅ Deployment complete!
echo.
echo To check status:
echo   kubectl get all -n kickoff
echo.
echo To access the application:
echo   http://localhost:8080/api/teams
echo.
pause
