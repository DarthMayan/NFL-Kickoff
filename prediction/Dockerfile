# Usar imagen base de Go
FROM golang:1.25-alpine AS builder

# Instalar dependencias necesarias
RUN apk add --no-cache git

# Establecer directorio de trabajo
WORKDIR /app

# Copiar archivos de módulo Go
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar el servicio prediction
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./prediction/cmd/main

# Etapa final - imagen mínima
FROM alpine:latest

# Instalar certificados SSL
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar el binario compilado
COPY --from=builder /app/main .

# Exponer puerto 8083
EXPOSE 8083

# Comando para ejecutar el servicio
CMD ["./main"]