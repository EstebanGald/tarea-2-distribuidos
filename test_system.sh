#!/bin/bash

# Script de testing para CyberDay Distribuido
# Verifica el funcionamiento correcto del sistema

set -e

echo "=========================================="
echo "  CyberDay Distribuido - Tests"
echo "=========================================="
echo ""

# Detectar comando de docker compose
if command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null 2>&1; then
    COMPOSE_CMD="docker compose"
else
    echo "Error: Docker Compose no está instalado"
    exit 1
fi

# Colores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

test_status() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}[PASS]${NC} $1"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}[FAIL]${NC} $1"
        ((TESTS_FAILED++))
    fi
}

echo "1. Verificando que todos los contenedores estén corriendo..."
docker-compose ps | grep -q "cyberday_broker.*Up"
test_status "Broker activo"

docker-compose ps | grep -q "cyberday_db1.*Up"
test_status "DB1 activo"

docker-compose ps | grep -q "cyberday_db2.*Up"
test_status "DB2 activo"

docker-compose ps | grep -q "cyberday_db3.*Up"
test_status "DB3 activo"

docker-compose ps | grep -q "cyberday_riploy"
test_status "Riploy ejecutado"

docker-compose ps | grep -q "cyberday_falabellox"
test_status "Falabellox ejecutado"

docker-compose ps | grep -q "cyberday_parisio"
test_status "Parisio ejecutado"

echo ""
echo "2. Verificando conectividad entre servicios..."

# Test broker -> db1
docker exec cyberday_broker nc -zv db1 50052 2>&1 | grep -q "succeeded"
test_status "Broker puede conectar a DB1"

# Test broker -> db2
docker exec cyberday_broker nc -zv db2 50053 2>&1 | grep -q "succeeded"
test_status "Broker puede conectar a DB2"

# Test broker -> db3
docker exec cyberday_broker nc -zv db3 50054 2>&1 | grep -q "succeeded"
test_status "Broker puede conectar a DB3"

echo ""
echo "3. Verificando logs de productores..."

docker-compose logs riploy | grep -q "Oferta.*enviada"
test_status "Riploy envió ofertas"

docker-compose logs falabellox | grep -q "Oferta.*enviada"
test_status "Falabellox envió ofertas"

docker-compose logs parisio | grep -q "Oferta.*enviada"
test_status "Parisio envió ofertas"

echo ""
echo "4. Verificando logs del broker..."

docker-compose logs broker | grep -q "Recibida oferta"
test_status "Broker recibió ofertas"

docker-compose logs broker | grep -q "almacenada con.*confirmaciones"
test_status "Broker almacenó ofertas en DB (W=2)"

docker-compose logs broker | grep -q "Registrando consumidor"
test_status "Broker registró consumidores"

echo ""
echo "5. Verificando logs de nodos DB..."

docker-compose logs db1 | grep -q "Guardando oferta"
test_status "DB1 guardó ofertas"

docker-compose logs db2 | grep -q "Guardando oferta"
test_status "DB2 guardó ofertas"

docker-compose logs db3 | grep -q "Guardando oferta"
test_status "DB3 guardó ofertas"

echo ""
echo "6. Verificando logs de consumidores..."

docker-compose logs consumidor_e1 | grep -q "Recibida oferta"
test_status "Consumidor E1 recibió ofertas"

docker-compose logs consumidor_h4 | grep -q "Recibida oferta"
test_status "Consumidor H4 (sin filtros) recibió ofertas"

echo ""
echo "7. Verificando simulación de fallos..."

docker-compose logs db2 | grep -q "SIMULANDO FALLO"
test_status "DB2 simuló fallo"

docker-compose logs db2 | grep -q "RECUPERADO DE FALLO"
test_status "DB2 se recuperó del fallo"

docker-compose logs consumidor_e3 | grep -q "SIMULANDO DESCONEXIÓN"
test_status "Consumidor E3 simuló desconexión"

docker-compose logs consumidor_e3 | grep -q "RECONECTADO"
test_status "Consumidor E3 se reconectó"

echo ""
echo "8. Verificando resincronización..."

docker-compose logs db2 | grep -q "Resincronizadas.*ofertas"
test_status "DB2 se resincronizó después del fallo"

docker-compose logs consumidor_e3 | grep -q "Solicitando histórico"
test_status "Consumidor E3 solicitó histórico después de reconectar"

echo ""
echo "9. Verificando archivos generados..."

# Verificar que se generaron CSVs de consumidores
docker exec cyberday_consumidor_e1 test -f /data/C-E1.csv
test_status "Consumidor E1 generó archivo CSV"

docker exec cyberday_consumidor_h4 test -f /data/C-H4.csv
test_status "Consumidor H4 generó archivo CSV"

echo ""
echo "10. Verificando persistencia en nodos DB..."

docker exec cyberday_db1 test -f /data/DB1_ofertas.json
test_status "DB1 persistió ofertas a disco"

docker exec cyberday_db2 test -f /data/DB2_ofertas.json
test_status "DB2 persistió ofertas a disco"

docker exec cyberday_db3 test -f /data/DB3_ofertas.json
test_status "DB3 persistió ofertas a disco"

echo ""
echo "=========================================="
echo "  Resumen de Tests"
echo "=========================================="
echo ""
echo -e "${GREEN}Tests pasados: $TESTS_PASSED${NC}"
echo -e "${RED}Tests fallidos: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ Todos los tests pasaron exitosamente${NC}"
    exit 0
else
    echo -e "${RED}❌ Algunos tests fallaron${NC}"
    echo ""
    echo "Sugerencias:"
    echo "  - Verifica los logs: make logs"
    echo "  - Verifica el estado: make status"
    echo "  - Reinicia servicios fallidos: make restart-[componente]"
    echo ""
    exit 1
fi