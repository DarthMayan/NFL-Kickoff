# Próximos Pasos - Migración a GORM

## ✅ Completado

1. **User Service** - Totalmente migrado a GORM
   - ✅ Modelos en `user/internal/models/`
   - ✅ Database layer en `user/internal/database/`
   - ✅ `main.go` actualizado para usar GORM
   - ✅ Todas las operaciones CRUD usando PostgreSQL

2. **Modelos Internos Creados**
   - ✅ `game/internal/models/` - Game y Team
   - ✅ `prediction/internal/models/` - Prediction
   - ✅ `leaderboard/internal/models/` - UserStats

3. **Capa de Base de Datos Creada**
   - ✅ Cada servicio tiene su `internal/database/database.go`

4. **teams.go Actualizado**
   - ✅ `game/internal/data/teams.go` ahora usa modelos internos

## 🔄 En Proceso

### Game Service
El archivo `game/cmd/main/main.go` necesita actualizarse similar al User Service.

**Patrón a seguir** (basado en User Service):
```go
// 1. Imports
import (
    "kickoff.com/game/internal/database"
    "kickoff.com/game/internal/models"
    "kickoff.com/game/internal/data"
    pb "kickoff.com/proto"
)

// 2. En main()
func main() {
    // ... puerto y flags ...

    // Conectar a DB
    if err := database.Connect(); err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer database.Close()

    // Cargar equipos NFL en la base de datos (solo primera vez)
    loadNFLTeams()

    // Cargar juegos de ejemplo
    loadSampleGames()

    // ... resto del código gRPC ...
}

// 3. Función helper para cargar teams
func loadNFLTeams() {
    for _, teamData := range data.NFLTeams {
        // Upsert para evitar duplicados
        var existing models.Team
        result := database.DB.Where("id = ?", teamData.ID).First(&existing)

        if result.Error != nil {
            // No existe, crear
            if err := database.DB.Create(&teamData).Error; err != nil {
                log.Printf("Error creating team %s: %v", teamData.ID, err)
            } else {
                log.Printf("Created team: %s", teamData.Name)
            }
        } else {
            log.Printf("Team already exists: %s", teamData.Name)
        }
    }
}

// 4. Función helper para cargar juegos
func loadSampleGames() {
    sampleGames := []models.Game{
        {
            ID:         "game_1",
            Week:       1,
            Season:     2024,
            HomeTeamID: "KC",
            AwayTeamID: "SF",
            GameTime:   time.Now().Add(24 * time.Hour),
            Status:     models.GameStatusScheduled,
        },
        // ... más juegos ...
    }

    for _, game := range sampleGames {
        var existing models.Game
        result := database.DB.Where("id = ?", game.ID).First(&existing)
        if result.Error != nil {
            database.DB.Create(&game)
            log.Printf("Created game: %s", game.ID)
        }
    }
}

// 5. Métodos gRPC - Reemplazar maps por GORM queries
func (gs *GameService) GetAllTeams(ctx context.Context, req *pb.GetAllTeamsRequest) (*pb.GetAllTeamsResponse, error) {
    var teams []models.Team
    if err := database.DB.Find(&teams).Error; err != nil {
        return nil, status.Errorf(codes.Internal, "failed to fetch teams: %v", err)
    }

    var pbTeams []*pb.Team
    for _, team := range teams {
        pbTeams = append(pbTeams, modelTeamToProto(team))
    }

    return &pb.GetAllTeamsResponse{
        Teams: pbTeams,
        Total: int32(len(pbTeams)),
    }, nil
}
```

**Cambios Clave**:
- Eliminar `teams map[string]models.Team` y `games map[string]models.Game`
- Eliminar `mu sync.RWMutex` y `counter int`
- Reemplazar todos los accesos a maps con queries GORM
- `database.DB.Find()`, `database.DB.Where()`, `database.DB.Create()`, etc.

## ⏭️ Servicios Pendientes

### Prediction Service
Archivo: `prediction/cmd/main/main.go`

**Cambios necesarios**:
```go
import (
    "kickoff.com/prediction/internal/database"
    "kickoff.com/prediction/internal/models"
)

// En main()
database.Connect()
defer database.Close()

// Métodos gRPC usan database.DB en lugar de maps
```

### Leaderboard Service
Archivo: `leaderboard/cmd/main/main.go`

**Cambios necesarios**:
```go
import (
    "kickoff.com/leaderboard/internal/database"
    "kickoff.com/leaderboard/internal/models"
)

// En main()
database.Connect()
defer database.Close()

// Métodos gRPC usan database.DB
```

## 📦 Reconstruir y Redesplegar

Una vez actualizados todos los servicios:

### 1. Construir imágenes Docker
```bash
docker build -t kickoff-user-service:latest -f user/Dockerfile .
docker build -t kickoff-game-service:latest -f game/Dockerfile .
docker build -t kickoff-prediction-service:latest -f prediction/Dockerfile .
docker build -t kickoff-leaderboard-service:latest -f leaderboard/Dockerfile .
docker build -t kickoff-gateway-service:latest -f gateway/Dockerfile .
```

### 2. Cargar a Kind
```bash
kind load docker-image kickoff-user-service:latest --name kickoff
kind load docker-image kickoff-game-service:latest --name kickoff
kind load docker-image kickoff-prediction-service:latest --name kickoff
kind load docker-image kickoff-leaderboard-service:latest --name kickoff
kind load docker-image kickoff-gateway-service:latest --name kickoff
```

### 3. Redesplegar
```bash
# Eliminar pods existentes para forzar recreación
kubectl delete pods -n kickoff -l tier=backend

# O hacer rollout restart
kubectl rollout restart -n kickoff deployment/user-service
kubectl rollout restart -n kickoff deployment/game-service
kubectl rollout restart -n kickoff deployment/prediction-service
kubectl rollout restart -n kickoff deployment/leaderboard-service
```

### 4. Verificar logs
```bash
kubectl logs -n kickoff -l app=user --tail=50
kubectl logs -n kickoff -l app=game --tail=50
kubectl logs -n kickoff -l app=prediction --tail=50
kubectl logs -n kickoff -l app=leaderboard --tail=50
kubectl logs -n kickoff -l app=postgres --tail=50
```

## 🗑️ Limpieza Final

Después de verificar que todo funciona:

```bash
# Eliminar pkg/models (ya no se usa)
rm -rf pkg/models

# Confirmar que no hay referencias a pkg/models
grep -r "kickoff.com/pkg/models" --include="*.go"
```

## 🧪 Pruebas

### Verificar que las bases de datos se crearon
```bash
kubectl exec -it -n kickoff <postgres-pod> -- psql -U kickoff_user -c "\l"
```

Deberías ver:
- user_db
- game_db
- prediction_db
- leaderboard_db

### Verificar que las tablas se crearon
```bash
kubectl exec -it -n kickoff <postgres-pod> -- psql -U kickoff_user -d user_db -c "\dt"
kubectl exec -it -n kickoff <postgres-pod> -- psql -U kickoff_user -d game_db -c "\dt"
```

### Probar endpoints
```bash
# Crear un usuario
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","fullName":"Test User"}'

# Obtener equipos
curl http://localhost:8080/api/teams

# Obtener juegos
curl http://localhost:8080/api/games
```

## 📝 Checklist Completo

- [x] User Service migrado a GORM
- [ ] Game Service migrado a GORM
- [ ] Prediction Service migrado a GORM
- [ ] Leaderboard Service migrado a GORM
- [x] teams.go actualizado
- [ ] Imágenes Docker reconstruidas
- [ ] Servicios redesplegados
- [ ] pkg/models eliminado
- [ ] Pruebas funcionales verificadas
- [ ] Frontend muestra datos de PostgreSQL

## 💡 Consejos

1. **Actualiza un servicio a la vez** y prueba antes de continuar
2. **Revisa los logs** después de cada deployment
3. **No elimines pkg/models** hasta que todos los servicios estén migrados
4. **Usa el User Service como referencia** - es el ejemplo completo
5. **Las bases de datos se auto-crean** en el primer arranque gracias al script init.sql

## 🐛 Troubleshooting

### Error: "relation does not exist"
- GORM automigrate debería crear las tablas
- Verifica logs del servicio para ver si la migración falló
- Verifica que DB_NAME esté configurado correctamente en el deployment

### Error: "connection refused"
- Verifica que postgres esté corriendo: `kubectl get pods -n kickoff`
- Verifica que el ConfigMap tenga las credenciales correctas

### Los datos no persisten
- Verifica que el PVC esté bound: `kubectl get pvc -n kickoff`
- Los datos se guardarán en el volumen persistente

### El servicio no arranca
- Ver logs: `kubectl logs -n kickoff <pod-name>`
- Común: error de compilación si falta algún import
