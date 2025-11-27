# Kubernetes Deployment - Kickoff NFL Prediction System

## 📋 Descripción

Este directorio contiene todos los manifiestos de Kubernetes necesarios para desplegar el sistema de predicciones NFL en un cluster de Kubernetes.

## 🏗️ Arquitectura

```
Internet
    ↓
LoadBalancer (Gateway Service)
    ↓
ClusterIP Services (Internal gRPC Communication)
    ├─ User Service (2-8 pods)
    ├─ Game Service (2-8 pods)
    ├─ Prediction Service (3-15 pods)
    └─ Leaderboard Service (2-10 pods)
```

## 📁 Estructura de Archivos

```
k8s/
├── base/
│   ├── namespace.yaml           # Namespace "kickoff"
│   └── hpa.yaml                 # HorizontalPodAutoscalers para todos los servicios
├── config/
│   └── configmap.yaml           # Configuraciones compartidas
├── deployments/
│   ├── gateway-deployment.yaml
│   ├── user-deployment.yaml
│   ├── game-deployment.yaml
│   ├── prediction-deployment.yaml
│   └── leaderboard-deployment.yaml
├── services/
│   ├── gateway-service.yaml     # LoadBalancer (acceso externo)
│   ├── user-service.yaml        # ClusterIP (interno)
│   ├── game-service.yaml        # ClusterIP (interno)
│   ├── prediction-service.yaml  # ClusterIP (interno)
│   └── leaderboard-service.yaml # ClusterIP (interno)
├── deploy.bat                   # Script de deployment para Windows
├── undeploy.bat                 # Script para eliminar todos los recursos
└── README.md                    # Este archivo
```

## 🚀 Deployment

### Prerequisitos

1. **Docker Desktop** con Kubernetes habilitado
2. **kubectl** instalado y configurado
3. **Imágenes Docker** construidas localmente

### Pasos para Desplegar

#### Opción 1: Script Automático (Recomendado)

```bash
# Ejecutar el script de deployment
.\k8s\deploy.bat
```

#### Opción 2: Deployment Manual

```bash
# 1. Construir imágenes Docker
docker-compose build

# 2. Crear namespace
kubectl apply -f k8s/base/namespace.yaml

# 3. Crear ConfigMaps
kubectl apply -f k8s/config/configmap.yaml

# 4. Crear Services
kubectl apply -f k8s/services/

# 5. Crear Deployments
kubectl apply -f k8s/deployments/

# 6. Crear HorizontalPodAutoscalers
kubectl apply -f k8s/base/hpa.yaml
```

### Verificar el Deployment

```bash
# Ver todos los recursos
kubectl get all -n kickoff

# Ver pods
kubectl get pods -n kickoff

# Ver services
kubectl get svc -n kickoff

# Ver HPA status
kubectl get hpa -n kickoff

# Ver logs de un pod
kubectl logs -f <pod-name> -n kickoff

# Describir un pod
kubectl describe pod <pod-name> -n kickoff
```

## 🌐 Acceder a la Aplicación

### Obtener la IP del LoadBalancer

```bash
kubectl get svc gateway-service -n kickoff
```

Buscar el `EXTERNAL-IP` en la salida. En Docker Desktop será `localhost`.

### Endpoints Disponibles

- **Health Check**: `http://localhost:8080/health`
- **Teams**: `http://localhost:8080/api/teams`
- **Games**: `http://localhost:8080/api/games`
- **Users**: `http://localhost:8080/api/users`
- **Predictions**: `http://localhost:8080/api/predictions`
- **Leaderboard**: `http://localhost:8080/api/leaderboard`

## 📊 Horizontal Pod Autoscaling (HPA)

### Configuración de Autoescalado

| Servicio | Min Pods | Max Pods | CPU Target | Memory Target |
|----------|----------|----------|------------|---------------|
| Gateway | 2 | 10 | 70% | 80% |
| User | 2 | 8 | 70% | 80% |
| Game | 2 | 8 | 70% | 80% |
| **Prediction** | **3** | **15** | **60%** | **75%** |
| Leaderboard | 2 | 10 | 65% | 75% |

**Nota**: Prediction Service tiene más pods porque es el servicio más crítico.

### Monitorear Autoescalado

```bash
# Ver estado de HPA en tiempo real
kubectl get hpa -n kickoff -w

# Ver detalles de un HPA específico
kubectl describe hpa prediction-hpa -n kickoff
```

### Generar Carga para Probar HPA

```bash
# Usar el script de load testing (requiere implementación)
# Ver sección de Load Testing más abajo
```

## 🔍 Troubleshooting

### Pods no inician

```bash
# Ver eventos del namespace
kubectl get events -n kickoff --sort-by='.lastTimestamp'

# Ver logs de un pod que falla
kubectl logs <pod-name> -n kickoff

# Describir el pod para ver errores
kubectl describe pod <pod-name> -n kickoff
```

### Problemas de ImagePullBackOff

Las imágenes están configuradas con `imagePullPolicy: Never` para usar imágenes locales.

Si tienes problemas:
```bash
# 1. Verificar que las imágenes existen localmente
docker images | grep kickoff

# 2. Asegurarte que Docker Desktop usa el mismo daemon que kubectl
docker context use default
```

### Service no responde

```bash
# Verificar endpoints del service
kubectl get endpoints -n kickoff

# Port-forward para debugging
kubectl port-forward svc/gateway-service 8080:8080 -n kickoff
```

## 🗑️ Eliminar Deployment

### Opción 1: Script Automático

```bash
.\k8s\undeploy.bat
```

### Opción 2: Manual

```bash
# Eliminar namespace completo (elimina todos los recursos)
kubectl delete namespace kickoff
```

## 📈 Resource Limits

Cada servicio tiene definidos:

**Requests** (recursos garantizados):
- CPU: 100m (0.1 cores)
- Memory: 128Mi

**Limits** (recursos máximos):
- CPU: 500m (0.5 cores)
- Memory: 512Mi

## 🔒 Health Checks

Cada pod tiene configurado:

**Liveness Probe**:
- Endpoint: `/health`
- Initial Delay: 10s
- Period: 10s
- Timeout: 5s
- Failure Threshold: 3

**Readiness Probe**:
- Endpoint: `/health`
- Initial Delay: 5s
- Period: 5s
- Timeout: 3s
- Failure Threshold: 3

## 🎯 Próximos Pasos

### 1. Load Testing (Pendiente)

Crear scripts de carga para probar el HPA:
- Usar Locust o K6
- Generar tráfico HTTP hacia el Gateway
- Observar el autoescalado en acción

### 2. Base de Datos (Pendiente)

Actualmente los servicios usan datos en memoria. Para producción:
- Agregar PostgreSQL StatefulSet
- Migrar servicios a usar la DB
- Configurar PersistentVolumeClaims

### 3. Monitoring (Opcional)

- Prometheus para métricas
- Grafana para visualización
- Jaeger para distributed tracing

### 4. CI/CD (Opcional)

- GitHub Actions para build automático
- ArgoCD para GitOps
- Helm charts para templating

## 📝 Notas

- **Namespace**: Todos los recursos se crean en el namespace `kickoff`
- **Service Discovery**: En Kubernetes, los services se descubren via DNS (ej: `user-service.kickoff.svc.cluster.local`)
- **gRPC**: La comunicación interna entre servicios usa gRPC (puertos 90XX)
- **HTTP**: El Gateway expone HTTP al exterior (puerto 8080)

## 🆘 Soporte

Si encuentras problemas:

1. Verifica que Docker Desktop tiene Kubernetes habilitado
2. Verifica que tienes suficientes recursos (CPU/Memory)
3. Revisa los logs con `kubectl logs`
4. Verifica eventos con `kubectl get events -n kickoff`

## 📚 Referencias

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [HPA Documentation](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
