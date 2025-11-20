# 🏈 Kickoff - NFL Prediction Game

Una aplicación de microservicios distribuidos para predicciones NFL con Go, Docker y Consul.

## 🐳 **SOLO DOCKER** - Ejecución Simplificada

### ✅ Iniciar Proyecto Completo:
```bash
./start-project.bat
```

### ⛔ Detener Proyecto:
```bash
./stop-project.bat
```

## 🏗️ Arquitectura de Microservicios

- **Consul** (Puerto 8500) - Service Discovery y Health Checks
- **Gateway** (Puerto 8080) - API Gateway principal
- **User Service** (Puerto 8081) - Gestión de usuarios
- **Game Service** (Puerto 8082) - Equipos y juegos NFL
- **Prediction Service** (Puerto 8083) - Predicciones de usuarios
- **Leaderboard Service** (Puerto 8084) - Rankings y estadísticas

## 📋 Requisitos

- **Docker Desktop** (único requisito)
- Git

## 🚀 Configuración Instantánea

```bash
git clone <repository-url>
cd kickoff
./start-project.bat
```

**¡Listo!** Todos los servicios se ejecutan automáticamente.

## 🔗 URLs Principales

### Interfaces Web:
- **Consul UI**: http://localhost:8500 (Service Discovery)
- **Gateway**: http://localhost:8080/health (API Gateway principal)

### Health Checks:
- **Gateway**: http://localhost:8080/health
- **User Service**: http://localhost:8081/health
- **Game Service**: http://localhost:8082/health
- **Prediction Service**: http://localhost:8083/health
- **Leaderboard Service**: http://localhost:8084/health

### APIs a través del Gateway (Recomendado):
- **Usuarios**: http://localhost:8080/api/users
- **Equipos NFL**: http://localhost:8080/api/teams
- **Juegos NFL**: http://localhost:8080/api/games
- **Predicciones**: http://localhost:8080/api/predictions
- **Leaderboard**: http://localhost:8080/api/leaderboard
- **Estadísticas Usuario**: http://localhost:8080/api/user-stats/{userID}
- **Predicciones Usuario**: http://localhost:8080/api/predictions/user/{userID}

### APIs Directas (Servicios individuales):
- **Usuarios**: http://localhost:8081/v2/users
- **Equipos NFL**: http://localhost:8082/v2/teams
- **Predicciones**: http://localhost:8083/v2/predictions

## 📖 Uso de APIs

### 🌐 A través del Gateway (Recomendado):

#### 1. Ver usuarios:
```bash
curl http://localhost:8080/api/users
```

#### 2. Ver equipos NFL:
```bash
curl http://localhost:8080/api/teams
```

#### 3. Ver juegos NFL:
```bash
curl http://localhost:8080/api/games
```

#### 4. Ver predicciones:
```bash
curl http://localhost:8080/api/predictions
```

#### 5. Ver leaderboard:
```bash
curl http://localhost:8080/api/leaderboard
```

#### 6. Ver estadísticas de usuario específico:
```bash
curl http://localhost:8080/api/user-stats/user_1
```

#### 7. Ver predicciones de usuario específico:
```bash
curl http://localhost:8080/api/predictions/user/user_1
```

### 🔧 Directamente a servicios individuales:

#### 1. Crear usuario:
```bash
curl -X POST http://localhost:8081/v2/users \
  -H "Content-Type: application/json" \
  -d '{"username": "john_doe", "email": "john@example.com", "fullName": "John Doe"}'
```

#### 2. Ver equipos:
```bash
curl http://localhost:8082/v2/teams
```

#### 3. Crear predicción:
```bash
curl -X POST http://localhost:8083/v2/predictions \
  -H "Content-Type: application/json" \
  -d '{"userId": "user_1", "gameId": "1", "predictedWinnerId": "KC"}'
```

## 🛠️ Monitoreo

- **Estado servicios**: `docker-compose ps`
- **Logs en tiempo real**: `docker-compose logs -f`
- **Logs servicio específico**: `docker-compose logs -f user-service`

## 🏗️ Arquitectura Técnica

- **Lenguaje**: Go
- **Service Discovery**: Consul
- **Contenedores**: Docker + Docker Compose
- **Comunicación**: APIs REST
- **Patrones**: Microservicios, Clean Architecture
- **Storage**: In-memory (desarrollo)

## ⚠️ Solución de Problemas

### Error de puerto en uso:
```bash
# Verificar qué está usando los puertos
netstat -an | findstr ":8500"
netstat -an | findstr ":8080"

# Detener todos los contenedores
docker-compose down --volumes --remove-orphans
```

### Los servicios no inician:
```bash
# Ver logs de servicios
docker-compose logs consul
docker-compose logs user-service

# Limpiar y reconstruir
docker-compose down --volumes --remove-orphans
docker-compose up --build -d
```

### Comandos útiles:
```bash
# Ver estado de servicios
docker-compose ps

# Ver logs en tiempo real
docker-compose logs -f

# Parar todo y limpiar
./stop-project.bat
```

## 📚 Scripts Incluidos

- `start-project.bat`: Inicia todo el proyecto automáticamente
- `stop-project.bat`: Detiene y limpia todos los contenedores

---

**✅ Proyecto optimizado para desarrollo y aprendizaje de sistemas distribuidos con Docker.**