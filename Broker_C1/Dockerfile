# Broker_C1/Dockerfile
FROM golang:1.24-alpine AS builder

# Instalar dependencias de compilación
RUN apk add --no-cache git

WORKDIR /app

# Copiar archivos de módulo primero (para aprovechar cache de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente completo
COPY . .

# Compilar el broker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o broker ./Broker/broker_main.go

# Imagen final minimal
FROM alpine:latest

# Instalar certificados CA y timezone data
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /root/

# Copiar binario compilado desde builder
COPY --from=builder /app/broker .

# Cambiar permisos
RUN chown appuser:appgroup /root/broker

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 50051

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD nc -z localhost 50051 || exit 1

# Ejecutar broker
CMD ["./broker"]