# Falabellox_BD2_C3/Falabellox/Dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY go.mod go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY proto ./proto

# Copiar código fuente y catálogo del productor
COPY Falabellox/falabellox.go ./Falabellox/
COPY Falabellox/falabellox_catalogo.csv ./Falabellox/

# Compilar desde el directorio correcto
WORKDIR /app/Falabellox
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o productor falabellox.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/Falabellox/productor .

# Copiar catálogo CSV
COPY --from=builder /app/Falabellox/falabellox_catalogo.csv .

RUN chown -R appuser:appgroup /app

USER appuser

ENV PRODUCTOR_NOMBRE=Falabellox
ENV CATALOGO=falabellox_catalogo.csv

CMD ["./productor"]
