# Falabellox_BD2_C3/Falabellox/Dockerfile
FROM golang:1.24-alpine AS builder

# Instalar dependencias
RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY ../go.mod ../go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY ../proto ./proto

# Copiar código fuente y catálogo del productor
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o productor ./falabellox.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/productor .

# Copiar catálogo CSV
COPY --from=builder /app/falabellox_catalogo.csv .

# Cambiar permisos
RUN chown -R appuser:appgroup /app

# Cambiar a usuario no-root
USER appuser

# Variables de entorno por defecto
ENV PRODUCTOR_NOMBRE=Falabellox
ENV CATALOGO=falabellox_catalogo.csv

# Ejecutar
CMD ["./productor"]