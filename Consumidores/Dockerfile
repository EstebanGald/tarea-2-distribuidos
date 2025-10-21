# Consumidores/Dockerfile
FROM golang:1.24-alpine AS builder

# Instalar dependencias
RUN apk add --no-cache git

WORKDIR /app

# Copiar archivos de módulo
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente completo
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o consumidor ./consumidor.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Crear usuario no-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Crear directorios
RUN mkdir -p /app /data && chown -R appuser:appgroup /app /data

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/consumidor .

# Cambiar permisos
RUN chown appuser:appgroup /app/consumidor

# Cambiar a usuario no-root
USER appuser

# Cambiar directorio de trabajo a /data para guardar CSVs
WORKDIR /data

# Healthcheck (varía según puerto del consumidor)
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD nc -z localhost 50061 || nc -z localhost 50062 || nc -z localhost 50063 || exit 1

# Variables de entorno por defecto
ENV CONSUMIDOR_ID=C-E1
ENV BROKER_ADDR=broker:50051
ENV ARCHIVO_CONFIG=/app/consumidores.csv

# Ejecutar consumidor
CMD ["/app/consumidor"]