# Falabellox_BD2_C3/BD2/Dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY go.mod go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY proto ./proto

# Copiar código fuente del nodo BD2
COPY BD2/bd2.go ./BD2/

# Compilar desde el directorio correcto
WORKDIR /app/BD2
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o db_node bd2.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN mkdir -p /data && chown appuser:appgroup /data

WORKDIR /data

# Copiar binario desde builder
COPY --from=builder /app/BD2/db_node .

RUN chown appuser:appgroup /data/db_node

USER appuser

EXPOSE 50053

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD nc -z localhost 50053 || exit 1

ENV NODO_ID=DB2
ENV PUERTO=:50053

CMD ["./db_node"]
