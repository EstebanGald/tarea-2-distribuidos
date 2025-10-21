# Parisio_BD3/BD3/Dockerfile
FROM golang:1.24-alpine AS builder

# Instalar dependencias
RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY ../go.mod ../go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY ../proto ./proto

# Copiar código fuente del nodo
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o db_node ./bd3.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Crear usuario no-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Crear directorio de datos
RUN mkdir -p /data && chown appuser:appgroup /data

WORKDIR /data

# Copiar binario
COPY --from=builder /app/db_node .

# Cambiar permisos
RUN chown appuser:appgroup /data/db_node

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 50054

# Healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD nc -z localhost 50054 || exit 1

# Variables de entorno por defecto
ENV NODO_ID=DB3
ENV PUERTO=:50054

# Ejecutar
CMD ["./db_node"]