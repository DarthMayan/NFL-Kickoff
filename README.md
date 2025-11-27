# Kickoff - Sistema de Predicciones NFL

Sistema de microservicios para predicciones de juegos de la NFL, implementado con Kubernetes, gRPC y PostgreSQL.

## 📋 Arquitectura del Proyecto

### Microservicios
1. **Gateway Service** - API HTTP que expone endpoints REST (LoadBalancer)
2. **User Service** - Gestión de usuarios (ClusterIP)
3. **Game Service** - Gestión de equipos y juegos NFL (ClusterIP)
4. **Prediction Service** - Gestión de predicciones (ClusterIP)
5. **Leaderboard Service** - Rankings y estadísticas (ClusterIP)

### Base de Datos
- **PostgreSQL** - Con bases de datos separadas por servicio:
  - `user_db` - User Service
  - `game_db` - Game Service
  - `prediction_db` - Prediction Service
  - `leaderboard_db` - Leaderboard Service

### Comunicación
- **Frontend → Gateway**: HTTP/REST
- **Gateway → Services**: gRPC
- **Services → PostgreSQL**: SQL via GORM

## 🔧 Requisitos Cumplidos

✅ Clúster de Kubernetes con mínimo 3 microservicios comunicados via gRPC
✅ Un microservicio expuesto al exterior vía LoadBalancer (Gateway)
✅ Microservicios internos comunicados por ClusterIP
✅ Base de datos PostgreSQL con Service tipo ClusterIP
✅ Cada servicio con su propia base de datos

## 🚀 Deployment en Kind (Kubernetes in Docker)

### 1. Crear el Cluster Kind

```bash
# Crear cluster con configuración especial para LoadBalancer
kind create cluster --name kickoff --config kind-config.yaml
```

### 2. Construir Imágenes Docker

```bash
# Construir imágenes de todos los servicios
docker build -t kickoff-gateway-service:latest -f gateway/Dockerfile .
docker build -t kickoff-user-service:latest -f user/Dockerfile .
docker build -t kickoff-game-service:latest -f game/Dockerfile .
docker build -t kickoff-prediction-service:latest -f prediction/Dockerfile .
docker build -t kickoff-leaderboard-service:latest -f leaderboard/Dockerfile .
```

### 3. Cargar Imágenes en Kind

```bash
# Cargar imágenes al cluster de Kind
kind load docker-image kickoff-gateway-service:latest --name kickoff
kind load docker-image kickoff-user-service:latest --name kickoff
kind load docker-image kickoff-game-service:latest --name kickoff
kind load docker-image kickoff-prediction-service:latest --name kickoff
kind load docker-image kickoff-leaderboard-service:latest --name kickoff
```

O usar el script:
```bash
./load-images-to-kind.bat
```

### 4. Desplegar en Kubernetes

```bash
# Crear namespace
kubectl apply -f k8s/base/namespace.yaml

# Aplicar ConfigMaps y configuración
kubectl apply -f k8s/config/

# Crear PersistentVolumeClaim para PostgreSQL
kubectl apply -f k8s/base/postgres-pvc.yaml

# Desplegar servicios
kubectl apply -f k8s/deployments/
kubectl apply -f k8s/services/
```

O usar el Makefile:
```bash
make deploy
```

### 5. Verificar Deployment

```bash
# Ver estado de los pods
kubectl get pods -n kickoff

# Ver servicios
kubectl get services -n kickoff

# Ver logs del Gateway
kubectl logs -n kickoff -l app=gateway

# Ver logs de PostgreSQL
kubectl logs -n kickoff -l app=postgres
```

### 6. Acceder al Frontend

El frontend se encuentra en `frontend/gateway-client/index.html`. Simplemente ábrelo en un navegador.

**URL del Gateway**: `http://localhost:8080`

Kind mapea el LoadBalancer del Gateway al puerto 8080 del host (configurado en `kind-config.yaml`).

## 📁 Estructura del Proyecto

```
kickoff/
├── gateway/              # API Gateway (HTTP → gRPC)
│   ├── cmd/main/
│   └── Dockerfile
├── user/                 # Servicio de Usuarios
│   ├── cmd/main/
│   ├── internal/
│   │   ├── models/      # Modelos GORM
│   │   └── database/    # Conexión DB
│   └── Dockerfile
├── game/                 # Servicio de Juegos
│   ├── cmd/main/
│   ├── internal/
│   │   ├── models/
│   │   ├── database/
│   │   └── data/        # Datos NFL
│   └── Dockerfile
├── prediction/           # Servicio de Predicciones
│   ├── cmd/main/
│   ├── internal/
│   │   ├── models/
│   │   └── database/
│   └── Dockerfile
├── leaderboard/          # Servicio de Leaderboard
│   ├── cmd/main/
│   ├── internal/
│   │   ├── models/
│   │   └── database/
│   └── Dockerfile
├── proto/                # Definiciones gRPC
├── k8s/                  # Manifiestos Kubernetes
│   ├── base/            # Namespace, PVC
│   ├── config/          # ConfigMaps
│   ├── deployments/     # Deployments
│   └── services/        # Services
├── frontend/            # Cliente web
│   └── gateway-client/
├── db/                  # Scripts SQL
└── kind-config.yaml     # Configuración Kind
```

## 🛠️ Comandos Útiles

### Desarrollo

```bash
# Regenerar código gRPC
./generate-proto.bat

# Compilar un servicio localmente
cd user && go build -o user.exe ./cmd/main

# Ejecutar tests
go test ./...
```

### Kubernetes

```bash
# Port-forward para acceder a PostgreSQL directamente
kubectl port-forward -n kickoff service/postgres-service 5432:5432

# Port-forward para acceder al Gateway
kubectl port-forward -n kickoff service/gateway-service 8080:8080

# Reiniciar un deployment
kubectl rollout restart -n kickoff deployment/user-service

# Ver eventos
kubectl get events -n kickoff --sort-by='.lastTimestamp'

# Ejecutar shell en un pod
kubectl exec -it -n kickoff <pod-name> -- /bin/sh

# Ver logs en tiempo real
kubectl logs -f -n kickoff <pod-name>
```

### Limpieza

```bash
# Eliminar todos los recursos del namespace
kubectl delete namespace kickoff

# O usar el Makefile
make clean

# Eliminar el cluster Kind
kind delete cluster --name kickoff
```

## 🔍 Testing

### Probar Endpoints del Gateway

```bash
# Health check
curl http://localhost:8080/health

# Obtener equipos
curl http://localhost:8080/api/teams

# Obtener juegos
curl http://localhost:8080/api/games

# Obtener usuarios
curl http://localhost:8080/api/users

# Obtener predicciones
curl http://localhost:8080/api/predictions

# Obtener leaderboard
curl http://localhost:8080/api/leaderboard
```

### Load Testing

Se incluyen scripts de pruebas de carga con k6:

```bash
# Test de carga básico
k6 run k6-load-test.js

# Test de estrés
k6 run k6-stress-test.js
```

## 📊 Monitoreo

### Ver Métricas de los Pods

```bash
# Uso de CPU y memoria
kubectl top pods -n kickoff

# Uso de nodos
kubectl top nodes
```

### HPA (Horizontal Pod Autoscaler)

El proyecto incluye configuración de HPA para escalar automáticamente:

```bash
# Ver HPA
kubectl get hpa -n kickoff

# Detalles del HPA
kubectl describe hpa <hpa-name> -n kickoff
```

## 🐛 Troubleshooting

### Los pods no arrancan

```bash
# Ver detalles del pod
kubectl describe pod -n kickoff <pod-name>

# Ver logs
kubectl logs -n kickoff <pod-name>

# Ver eventos
kubectl get events -n kickoff
```

### PostgreSQL no está listo

```bash
# Verificar que el PVC está bound
kubectl get pvc -n kickoff

# Ver logs de PostgreSQL
kubectl logs -n kickoff -l app=postgres

# Verificar que las bases de datos se crearon
kubectl exec -it -n kickoff <postgres-pod> -- psql -U kickoff_user -c "\l"
```

### Gateway no puede conectarse a los servicios

```bash
# Verificar que los servicios existen
kubectl get services -n kickoff

# Probar resolución DNS desde un pod
kubectl exec -it -n kickoff <gateway-pod> -- nslookup user-service

# Verificar que los puertos son correctos
kubectl describe service user-service -n kickoff
```

### Frontend no se conecta

1. Verificar que Kind está corriendo: `kind get clusters`
2. Verificar que el Gateway tiene LoadBalancer: `kubectl get svc -n kickoff gateway-service`
3. Verificar mapping de puertos en `kind-config.yaml`
4. Abrir DevTools del navegador y verificar errores de CORS o red

## 📝 Notas Importantes

### ⚠️ Estado Actual del Proyecto

**Los servicios actualmente usan almacenamiento en memoria** (maps). Para usar PostgreSQL con GORM:

1. Cada servicio necesita actualizar su `main.go` para:
   - Importar `internal/database` e `internal/models`
   - Llamar a `database.Connect()` al iniciar
   - Reemplazar operaciones con maps por GORM queries

2. Ver `MIGRATION_STATUS.md` para detalles de la migración a GORM

3. Los modelos y la capa de base de datos ya están creados en cada servicio

### Para Producción

- Usar Secrets en lugar de ConfigMaps para passwords
- Habilitar SSL/TLS para gRPC
- Implementar autenticación y autorización
- Configurar backups de PostgreSQL
- Usar un servicio de Load Balancer real (no Kind)
- Implementar observabilidad (Prometheus, Grafana, Jaeger)

## 👥 Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto es parte de un trabajo académico para el curso de Computación Distribuida.

## 🙏 Agradecimientos

- Kubernetes
- gRPC
- GORM
- PostgreSQL
- Kind (Kubernetes in Docker)
