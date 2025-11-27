# Estado de Migración a Kubernetes con GORM

## ✅ Cambios Completados

### 1. Arquitectura Kubernetes
- ✅ **Gateway Service**: Cambiado de NodePort a **LoadBalancer** (cumple requisitos)
- ✅ **Services Backend**: User, Game, Prediction, Leaderboard usan **ClusterIP** (cumple requisitos)
- ✅ **PostgreSQL Service**: Usa **ClusterIP** (cumple requisitos)
- ✅ **Comunicación gRPC**: Gateway se comunica con todos los servicios via gRPC

### 2. Base de Datos PostgreSQL
- ✅ **Múltiples bases de datos**: Creado script de inicialización para crear 4 bases de datos separadas:
  - `user_db` - Para User Service
  - `game_db` - Para Game Service
  - `prediction_db` - Para Prediction Service
  - `leaderboard_db` - Para Leaderboard Service
- ✅ **ConfigMaps**: Actualizado `postgres-config.yaml` y creado `postgres-init-script.yaml`
- ✅ **Deployment**: PostgreSQL configurado para ejecutar script de inicialización automáticamente

### 3. Modelos Separados por Servicio
- ✅ **User Service**: `user/internal/models/user.go` con modelo GORM
- ✅ **Game Service**: `game/internal/models/game.go` con modelos Team y Game con GORM
- ✅ **Prediction Service**: `prediction/internal/models/prediction.go` con modelo GORM
- ✅ **Leaderboard Service**: `leaderboard/internal/models/leaderboard.go` con modelo GORM

### 4. Capa de Base de Datos GORM
- ✅ **User Service**: `user/internal/database/database.go`
- ✅ **Game Service**: `game/internal/database/database.go`
- ✅ **Prediction Service**: `prediction/internal/database/database.go`
- ✅ **Leaderboard Service**: `leaderboard/internal/database/database.go`

### 5. Configuración de Deployments
- ✅ Todos los deployments actualizados con variables de entorno:
  - `DB_NAME` específico para cada servicio
  - ConfigMap `postgres-config` para credenciales compartidas

### 6. Dependencias
- ✅ `go.mod` actualizado con GORM y driver de PostgreSQL

## ⚠️ Cambios Pendientes

### 1. Actualizar Servicios para Usar Base de Datos
Cada servicio (`user`, `game`, `prediction`, `leaderboard`) necesita:
- Importar su paquete `internal/database` y `internal/models`
- Llamar a `database.Connect()` en el `main()`
- Reemplazar los mapas en memoria con consultas GORM
- Usar `database.DB` para operaciones CRUD

**Ejemplo para User Service** (`user/cmd/main/main.go`):
```go
import (
    "kickoff.com/user/internal/database"
    "kickoff.com/user/internal/models"
)

func main() {
    // ... código existente ...

    // Conectar a la base de datos
    if err := database.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer database.Close()

    // ... resto del código ...
}

// En los métodos gRPC, usar database.DB en lugar de mapas:
func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
    user := models.User{
        ID:       uuid.New().String(),
        Username: req.Username,
        Email:    req.Email,
        FullName: req.FullName,
    }

    if err := database.DB.Create(&user).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
    }

    // ... convertir a proto y devolver ...
}
```

### 2. Eliminar Dependencia de pkg/models
- Los servicios actualmente importan `kickoff.com/pkg/models`
- Deben cambiar a sus modelos internos: `kickoff.com/user/internal/models`, etc.
- Eliminar `pkg/models/` cuando ya no se use

### 3. Cargar Datos Iniciales (Game Service)
El Game Service necesita cargar los equipos NFL en la base de datos al iniciar:
```go
func loadNFLTeams() error {
    for _, teamData := range data.NFLTeams {
        team := models.Team{
            ID:         teamData.ID,
            Name:       teamData.Name,
            City:       teamData.City,
            Conference: teamData.Conference,
            Division:   teamData.Division,
            LogoURL:    teamData.LogoURL,
            Stadium:    teamData.Stadium,
        }

        // Usar upsert para evitar duplicados
        database.DB.Where(models.Team{ID: team.ID}).FirstOrCreate(&team)
    }
    return nil
}
```

### 4. Frontend
El frontend necesita:
- Actualizar la URL de conexión para apuntar al LoadBalancer
- En Kind, usar: `http://localhost:8080` (puerto mapeado en kind-config.yaml)
- Verificar que los endpoints del Gateway devuelvan los datos correctamente

### 5. Limpieza de Archivos Obsoletos
Eliminar:
- ❌ `README.md` (marcado para eliminación en git)
- ❌ `docker-compose.yml` (ya no se usa Docker Compose)
- ❌ `start-project.bat` y `stop-project.bat` (obsoletos)
- ❌ `pkg/discovery/` (Consul removido)
- ❌ Posiblemente `pkg/models/` después de migrar todos los servicios

## 📋 Resumen de Requisitos del Proyecto

### ✅ Cumplidos
1. ✅ Clúster de Kubernetes con mínimo 3 microservicios comunicados via gRPC
   - Gateway, User, Game, Prediction, Leaderboard (5 servicios)
2. ✅ Un microservicio expuesto al exterior via **LoadBalancer**
   - Gateway Service (tipo LoadBalancer)
3. ✅ Microservicios comunicados por **ClusterIP**
   - User, Game, Prediction, Leaderboard (todos ClusterIP)
4. ✅ Base de datos comunicada por **ClusterIP**
   - PostgreSQL Service (ClusterIP)

### ⚠️ En Proceso
- Cada microservicio debe interactuar con su propia base de datos usando GORM
- Separación completa de modelos (cada servicio con sus propios modelos)

## 🚀 Siguiente Pasos Recomendados

1. **Actualizar User Service** para usar GORM (prioridad alta)
2. **Actualizar Game Service** para usar GORM y cargar datos NFL
3. **Actualizar Prediction y Leaderboard Services**
4. **Probar deployment en Kind**:
   ```bash
   kubectl apply -f k8s/config/
   kubectl apply -f k8s/deployments/
   kubectl apply -f k8s/services/
   ```
5. **Verificar que los datos fluyen correctamente** desde frontend → gateway → servicios → postgres
6. **Limpieza final** de archivos obsoletos

## 📝 Notas Importantes

- **Service Discovery**: Kubernetes DNS automático (no necesita Consul)
- **Comunicación**: Gateway (HTTP) ↔ Services (gRPC) ↔ PostgreSQL
- **Escalabilidad**: Cada servicio tiene réplicas configuradas
- **Persistencia**: PostgreSQL con PersistentVolumeClaim

## 🔧 Comandos Útiles

```bash
# Crear el cluster Kind
kind create cluster --name kickoff --config kind-config.yaml

# Aplicar configuraciones
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/config/
kubectl apply -f k8s/base/postgres-pvc.yaml
kubectl apply -f k8s/deployments/
kubectl apply -f k8s/services/

# Ver estado
kubectl get pods -n kickoff
kubectl get services -n kickoff

# Ver logs
kubectl logs -n kickoff <pod-name>

# Acceder al Gateway desde el host
kubectl port-forward -n kickoff service/gateway-service 8080:8080
```
