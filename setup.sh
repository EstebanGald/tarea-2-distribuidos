#!/bin/bash

# Script de configuración inicial para CyberDay Distribuido
# Universidad Técnica Federico Santa María - Laboratorio 2

set -e

echo "=========================================="
echo "  CyberDay Distribuido - Setup"
echo "=========================================="
echo ""

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para imprimir con color
print_status() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

# Verificar prerequisitos
echo "Verificando prerequisitos..."
echo ""

# Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    print_status "Docker instalado: $DOCKER_VERSION"
else
    print_error "Docker no está instalado"
    exit 1
fi

# Docker Compose (detectar versión)
echo ""
echo "Detectando Docker Compose..."
if command -v docker-compose &> /dev/null; then
    COMPOSE_VERSION=$(docker-compose --version | cut -d' ' -f4 | cut -d',' -f1)
    COMPOSE_CMD="docker-compose"
    print_status "Docker Compose (standalone) instalado: $COMPOSE_VERSION"
elif docker compose version &> /dev/null 2>&1; then
    COMPOSE_VERSION=$(docker compose version --short 2>/dev/null || echo "v2+")
    COMPOSE_CMD="docker compose"
    print_status "Docker Compose (plugin) instalado: $COMPOSE_VERSION"
else
    print_error "Docker Compose no está instalado"
    print_error "Instala con: sudo apt-get install docker-compose-plugin"
    exit 1
fi

echo ""
echo "Usando comando: $COMPOSE_CMD"

# Go (opcional)
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status "Go instalado: $GO_VERSION (opcional)"
else
    print_warning "Go no está instalado (solo necesario para desarrollo)"
fi

# Make (opcional)
if command -v make &> /dev/null; then
    print_status "Make instalado"
else
    print_warning "Make no está instalado (comandos alternativos disponibles)"
fi

echo ""
echo "=========================================="
echo "  Configurando estructura de directorios"
echo "=========================================="
echo ""

# Crear directorios necesarios
mkdir -p Consumidores/proto
mkdir -p resultados
mkdir -p shared

print_status "Directorios creados"

# Verificar archivos necesarios
echo ""
echo "=========================================="
echo "  Verificando archivos necesarios"
echo "=========================================="
echo ""

check_file() {
    if [ -f "$1" ]; then
        print_status "Encontrado: $1"
        return 0
    else
        print_error "Faltante: $1"
        return 1
    fi
}

FILES_OK=true

check_file "consumidores.csv" || FILES_OK=false
check_file "docker-compose.yml" || FILES_OK=false

# Verificar catálogos
check_file "Riploy_BD1_C2/Riploy/riploy_catalogo.csv" || FILES_OK=false
check_file "Falabellox_BD2_C3/Falabellox/falabellox_catalogo.csv" || FILES_OK=false
check_file "Parisio_BD3/Parisio/parisio_catalogo.csv" || FILES_OK=false

if [ "$FILES_OK" = false ]; then
    echo ""
    print_error "Faltan archivos necesarios. Por favor verifica la estructura del proyecto."
    exit 1
fi

echo ""
echo "=========================================="
echo "  Recompilando Protocol Buffers"
echo "=========================================="
echo ""

if command -v protoc &> /dev/null; then
    # Recompilar proto files si protoc está disponible
    for dir in Broker_C1 Riploy_BD1_C2 Falabellox_BD2_C3 Parisio_BD3 Consumidores; do
        if [ -f "$dir/proto/ofertas.proto" ]; then
            echo "Compilando $dir/proto/ofertas.proto..."
            (cd "$dir" && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto 2>/dev/null) || print_warning "Error compilando en $dir (puede ser normal)"
        fi
    done
    print_status "Protocol Buffers procesados"
else
    print_warning "protoc no instalado, usando archivos .pb.go existentes"
fi

echo ""
echo "=========================================="
echo "  Construyendo imágenes Docker"
echo "=========================================="
echo ""

echo "Ejecutando: $COMPOSE_CMD build"
if $COMPOSE_CMD build; then
    print_status "Imágenes Docker construidas exitosamente"
else
    print_error "Error construyendo imágenes Docker"
    exit 1
fi

echo ""
echo "=========================================="
echo "  ✅ Setup Completado"
echo "=========================================="
echo ""
echo "Comandos disponibles:"
echo ""
echo "  Iniciar sistema completo:"
echo "    $COMPOSE_CMD up -d"
echo "    (o: make up si tienes Make)"
echo ""
echo "  Ver logs:"
echo "    $COMPOSE_CMD logs -f"
echo ""
echo "  Ver estado:"
echo "    $COMPOSE_CMD ps"
echo ""
echo "  Detener sistema:"
echo "    $COMPOSE_CMD down"
echo ""
echo "  Extraer resultados:"
echo "    make extraer-resultados"
echo "    (o manualmente: docker cp ...)"
echo ""
echo "=========================================="
echo ""
echo "💡 Tip: Guarda este comando para uso futuro:"
echo "    export COMPOSE_CMD='$COMPOSE_CMD'"
echo ""