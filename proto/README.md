# gRPC Proto Definitions

Este directorio contiene las definiciones de Protocol Buffers para la comunicación gRPC entre microservicios.

## Prerequisitos

### 1. Instalar Protocol Buffers Compiler (protoc)

**Windows:**
```bash
# Descargar desde: https://github.com/protocolbuffers/protobuf/releases
# O usar chocolatey:
choco install protoc

# Verificar instalación
protoc --version
```

**Linux/Mac:**
```bash
# Ubuntu/Debian
sudo apt install -y protobuf-compiler

# Mac
brew install protobuf

# Verificar instalación
protoc --version
```

### 2. Instalar plugins de Go para protoc

```bash
make proto-install
```

O manualmente:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Asegúrate de que `$GOPATH/bin` esté en tu `PATH`:
```bash
# Agregar a ~/.bashrc o ~/.zshrc
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Generar código Go desde archivos .proto

### Opción 1: Usar Makefile (recomendado)

```bash
# Generar código para todos los servicios
make proto-gen-all

# O específicamente para prediction service
make proto-gen
```

### Opción 2: Comando manual

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/prediction_service.proto
```

## Archivos generados

Después de ejecutar la generación, se crearán:

- `prediction_service.pb.go` - Estructuras de mensajes
- `prediction_service_grpc.pb.go` - Cliente y servidor gRPC

**IMPORTANTE:** Estos archivos son generados automáticamente. NO los edites manualmente.

## Actualizar dependencias

```bash
go mod tidy
```

## Estructura del servicio Prediction

### Endpoints gRPC:

1. **CreatePrediction** - Crear nueva predicción
2. **GetPredictionByID** - Obtener predicción por ID
3. **GetUserPredictions** - Obtener predicciones de un usuario
4. **GetGamePredictions** - Obtener predicciones de un juego
5. **GetWeekPredictions** - Obtener predicciones por semana
6. **GetAllPredictions** - Listar todas las predicciones
7. **DeletePrediction** - Eliminar predicción (solo pending)
8. **UpdatePredictionStatus** - Actualizar estado (interno)

### Configuración de puertos:

- **HTTP**: 8083 (mantener para compatibilidad)
- **gRPC**: 9083 (nuevo)

## Próximos pasos

1. Generar código: `make proto-gen`
2. Implementar servidor gRPC en prediction service
3. Actualizar gateway para usar cliente gRPC
4. Agregar más servicios (.proto para user, game, etc.)
