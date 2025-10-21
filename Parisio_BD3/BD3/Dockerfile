# Parisio_BD3/BD3/Dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY go.mod go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY proto ./proto

# Copiar código fuente del nodo BD3
COPY BD3/bd3.go ./BD3/

# Compilar desde el directorio correcto
WORKDIR /app/BD3
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o db_node bd3.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

RUN mkdir -p /data && chown appuser:appgroup /data

WORKDIR /data

# Copiar binario desde builder
COPY --from=builder /app/BD3/db_node .

RUN chown appuser:appgroup /data/db_node

USER appuser

EXPOSE 50054

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD nc -z localhost 50054 || exit 1

ENV NODO_ID=DB3
ENV PUERTO=:50054

CMD ["./db_node"]
