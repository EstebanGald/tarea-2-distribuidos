# Parisio_BD3/Parisio/Dockerfile
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copiar go.mod y go.sum desde el directorio padre
COPY go.mod go.sum ./
RUN go mod download

# Copiar proto desde directorio padre
COPY proto ./proto

# Copiar código fuente y catálogo del productor
COPY Parisio/parisio.go ./Parisio/
COPY Parisio/parisio_catalogo.csv ./Parisio/

# Compilar desde el directorio correcto
WORKDIR /app/Parisio
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o productor parisio.go

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copiar binario compilado
COPY --from=builder /app/Parisio/productor .

# Copiar catálogo CSV
COPY --from=builder /app/Parisio/parisio_catalogo.csv .

RUN chown -R appuser:appgroup /app

USER appuser

ENV PRODUCTOR_NOMBRE=Parisio
ENV CATALOGO=parisio_catalogo.csv

CMD ["./productor"]
