# Broker_C1/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fuente
COPY . .

# Compilar
RUN CGO_ENABLED=0 GOOS=linux go build -o broker ./Broker/broker_main.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar binario
COPY --from=builder /app/broker .

# Exponer puerto
EXPOSE 50051

# Ejecutar
CMD ["./broker"]