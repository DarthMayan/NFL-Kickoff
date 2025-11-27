# Kickoff NFL  Especificacion del Cluster

## 1. Componentes del cluster

| Servicio | Imagen | Replicas (base) | CPU (req/limit) | Memoria (req/limit) | Puerto | Funcion |
|----------|--------|-----------------|-----------------|---------------------|--------|---------|
| Gateway (`k8s/deployments/gateway-deployment.yaml`) | `kickoff-gateway-service:latest` | 2 | 100m / 500m | 128Mi / 512Mi | 8080 HTTP | Expone REST/HTTP, convierte a gRPC, sirve el frontend embebido y aplica CORS. |
| User Service (`k8s/deployments/user-deployment.yaml`) | `kickoff-user-service:latest` | 2 | 100m / 500m | 128Mi / 512Mi | 9081 gRPC | CRUD de usuarios, busqueda y validaciones; usa `user_db`. |
| Game Service (`k8s/deployments/game-deployment.yaml`) | `kickoff-game-service:latest` | 2 | 100m / 500m | 128Mi / 512Mi | 9082 gRPC | Catalogo NFL (teams/games), carga inicial de 32 equipos y 4 juegos. |
| Prediction Service (`k8s/deployments/prediction-deployment.yaml`) | `kickoff-prediction-service:latest` | 3 | 100m / 500m | 128Mi / 512Mi | 9083 gRPC | Registro/evaluacion de predicciones, puntuacion y estados. |
| Leaderboard Service (`k8s/deployments/leaderboard-deployment.yaml`) | `kickoff-leaderboard-service:latest` | 2 | 100m / 500m | 128Mi / 512Mi | 9084 gRPC | Agrega estadisticas, calcula leaderboard/precision y expone vistas de usuario. |
| PostgreSQL (`k8s/deployments/postgres-deployment.yaml`) | `postgres:15-alpine` | 1 | 100m / 500m | 128Mi / 512Mi | 5432 TCP | Motor relacional compartido con bases separadas por servicio y semillas. |

- **Comunicacion:** Gateway  servicios via gRPC (Kubernetes DNS), frontend  gateway via HTTP, servicios  PostgreSQL mediante GORM/GORM PostgreSQL driver.
- **Networking:** `k8s/services/*.yaml` define `ClusterIP` para servicios internos y un `LoadBalancer` (mapeado a NodePort 31479 en Kind) para el gateway.

## 2. Base de datos y esquema

- **Tipo:** PostgreSQL 15 (deployment + ConfigMaps en `k8s/config/postgres-config.yaml` y `k8s/config/postgres-init-script.yaml`).
- **Bases:** `db/init.sql` crea `user_db`, `game_db`, `prediction_db`, `leaderboard_db` y otorga privilegios al usuario `kickoff_user`.
- **Esquema:** `db/init-schema.sql` centraliza la definicion de tablas, indices, vistas y datos iniciales:
  - `users`, `teams`, `games`, `predictions` con llaves foraneas y constraints de integridad.
  - Disparador `update_predictions_updated_at` para mantener `updated_at`.
  - `leaderboard_view` para calculos agregados.
  - Seeds (32 equipos NFL, usuarios de prueba, juegos de ejemplo).
- **Persistencia:** En Kind se usa `emptyDir` para simplificar y regenerar datos en cada arranque; tambien se incluye `k8s/base/postgres-pvc.yaml` para escenarios con almacenamiento persistente.

## 3. Manifiestos Kubernetes (principales)

| Archivo | Descripcion |
|---------|-------------|
| `k8s/base/namespace.yaml` | Crea el namespace `kickoff`. |
| `k8s/base/postgres-pvc.yaml` | PVC de 1 Gi (`standard`) listo para entornos con almacenamiento persistente. |
| `k8s/base/hpa.yaml` | Define HPAs para gateway, user, game, prediction y leaderboard (min/max replicas, objetivos de CPU/Mem y politicas de escalado). |
| `k8s/config/configmap.yaml` | Variables comunes (DNS de servicios, `LOG_LEVEL`, `ENVIRONMENT`). |
| `k8s/config/postgres-config.yaml` | Credenciales/host/puerto compartidos para los microservicios. |
| `k8s/config/postgres-init-script.yaml` | Embebe `init.sql` para que el contenedor de PostgreSQL cree y otorgue las bases. |
| `k8s/deployments/*.yaml` | Especifican imagen, probes, recursos y replicas de cada microservicio y de PostgreSQL. |
| `k8s/services/*.yaml` | Publican puertos: gateway como `LoadBalancer`, el resto como `ClusterIP`. |

## 4. Extras

- **Frontend ligero:** `frontend/gateway-client/index.html` consume `/api/*` via fetch y se puede abrir desde el gateway o como HTML estatico.
- **Scripts operativos:** `deploy-to-kind.bat`, `load-images-to-kind.bat`, `create-kind-cluster.bat`, `run-k6-test.bat` agilizan CI/CD local.
- **Pruebas de carga:** `k6-load-test.js` y `k6-stress-test.js` generan 20-50 usuarios virtuales, con umbrales (`p95<500 ms`, `<5%` errores) para validar resiliencia.
- **Clientes gRPC:** `examples/*-grpc-client/main.go` facilitan smoke tests directos a cada servicio (utiles cuando el gateway aun no soportaba todas las operaciones).
- **Documentacion de migracion:** `MIGRATION_STATUS.md` y `NEXT_STEPS.md` registran la transicion de in-memory a PostgreSQL/GORM.

## 5. Capacidad y escalamiento

### 5.1 Recursos base (replicas actuales)

- **CPU solicitada total:** 1.2 vCPU (0.2 gateway + 0.2 user + 0.2 game + 0.3 prediction + 0.2 leaderboard + 0.1 postgres).
- **Limite total:** 6 vCPU (sumando todas las replicas) y 3 GiB de RAM (6 pods  512 MiB + postgres 512 MiB).
- **Uso tipico:** pensada para Kind (Docker Desktop con 6 vCPU/8 GiB disponibles).

### 5.2 Escalamiento horizontal

Segun `k8s/base/hpa.yaml`:

| Servicio | HPA (min-max) | Metrica | Capacidad maxima (CPU teorica) |
|----------|---------------|---------|--------------------------------|
| Gateway | 2-10 pods | 70% CPU / 80% RAM | 10 x 0.5 CPU = 5 vCPU (10k req/min si cada pod maneja 100 req/s). |
| User | 2-8 pods | 70% CPU / 80% RAM | 4 vCPU. |
| Game | 2-8 pods | 70% CPU / 80% RAM | 4 vCPU. |
| Prediction | 3-15 pods | 60% CPU / 75% RAM | 7.5 vCPU (servicio mas critico). |
| Leaderboard | 2-10 pods | 65% CPU / 75% RAM | 5 vCPU. |

En conjunto, el cluster puede escalar hasta **26 vCPU** y **26 GiB RAM** si todos los HPAs llegan a su maximo (util para cargas bursty).

### 5.3 Carga de trabajo soportada

- **Pruebas con K6 (`k6-load-test.js`):** el escenario por defecto mantiene 20 usuarios virtuales concurrentes durante 1 min. El escenario de rampa (comentado) sube a 50 VUs; el gateway, user, game y leaderboard mantuvieron `p95<500 ms` y <5 % de errores usando port-forward al gateway.
- **Interpretacion:** con 2 pods de gateway (1 vCPU total) y respuestas en <500 ms, se sostienen 40-50 solicitudes/seg. Al escalar a 10 pods (5 vCPU) el throughput esperado sube a 200-250 req/s, lo cual equivale a ~250 usuarios concurrentes navegando el frontend o ejecutando llamadas REST. Prediction Service (15 pods max) mantiene la cola de escritura con suficiente headroom.
- **Usuarios concurrentes:** conservadoramente, el sistema soporta **50 VUs** (validado) y esta dimensionado para llegar a **200-250 usuarios simultaneos** cuando los HPAs escalan (limitado por CPU disponible en el nodo Kind). Para entornos cloud con nodos adicionales se puede aumentar la capacidad agregando workers a Kind/Kubernetes.

## 6. Como probar el cluster

1. **Construir y cargar imagenes locales:**
   ```powershell
   docker build -t kickoff-<service>-service:latest -f <service>/Dockerfile .
   kind load docker-image kickoff-<service>-service:latest --name kickoff
   ```
2. **Aplicar manifiestos:**
   ```powershell
   kubectl apply -f k8s/base/namespace.yaml
   kubectl apply -n kickoff -f k8s/config
   kubectl apply -n kickoff -f k8s/deployments
   kubectl apply -n kickoff -f k8s/services
   kubectl apply -n kickoff -f k8s/base/hpa.yaml   # opcional durante demos
   ```
3. **Exponer gateway para el frontend y pruebas REST:**
   ```powershell
   kubectl port-forward -n kickoff service/gateway-service 8080:8080
   curl http://localhost:8080/health
   ```
4. **Ejecutar pruebas de carga opcionales:**
   ```powershell
   k6 run k6-load-test.js
   ```

Este documento resume el estado actual del cluster, sus dependencias y la capacidad estimada para respaldar la entrega del proyecto Kickoff NFL.
