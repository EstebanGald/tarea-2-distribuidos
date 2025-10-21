.PHONY: help build up down logs clean proto test

# Variables
# Detectar si usar docker-compose o docker compose
DOCKER_COMPOSE := $(shell if command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo "docker compose"; fi)

help:
	@echo "=== CyberDay Distribuido - Comandos Disponibles ==="
	@echo ""
	@echo "  make build          - Construir todas las imágenes Docker"
	@echo "  make up             - Iniciar todos los servicios"
	@echo "  make down           - Detener todos los servicios"
	@echo "  make logs           - Ver logs de todos los servicios"
	@echo "  make logs-broker    - Ver logs solo del broker"
	@echo "  make logs-db        - Ver logs de nodos DB"
	@echo "  make logs-prod      - Ver logs de productores"
	@echo "  make logs-cons      - Ver logs de consumidores"
	@echo "  make clean          - Limpiar contenedores y volúmenes"
	@echo "  make proto          - Recompilar archivos Protocol Buffers"
	@echo "  make test           - Ejecutar tests"
	@echo "  make docker-VM1     - Servicios para VM1"
	@echo "  make docker-VM2     - Servicios para VM2"
	@echo "  make docker-VM3     - Servicios para VM3"
	@echo "  make docker-VM4     - Servicios para VM4"
	@echo "  make reporte        - Mostrar reporte generado"
	@echo ""

# Construcción
build:
	@echo " Construyendo imágenes Docker..."
	$(DOCKER_COMPOSE) build

# Iniciar servicios
up:
	@echo " Iniciando CyberDay Distribuido..."
	$(DOCKER_COMPOSE) up -d
	@echo "✅ Sistema iniciado"
	@echo " Ver logs: make logs"

# Iniciar con logs visibles
up-logs:
	@echo " Iniciando CyberDay Distribuido con logs..."
	$(DOCKER_COMPOSE) up

# Detener servicios
down:
	@echo "🛑 Deteniendo servicios..."
	$(DOCKER_COMPOSE) down

# Ver logs
logs:
	$(DOCKER_COMPOSE) logs -f

logs-broker:
	$(DOCKER_COMPOSE) logs -f broker

logs-db:
	$(DOCKER_COMPOSE) logs -f db1 db2 db3

logs-prod:
	$(DOCKER_COMPOSE) logs -f riploy falabellox parisio

logs-cons:
	$(DOCKER_COMPOSE) logs -f consumidor_e1 consumidor_e2 consumidor_e3 consumidor_e4 \
		consumidor_m1 consumidor_m2 consumidor_m3 consumidor_m4 \
		consumidor_h1 consumidor_h2 consumidor_h3 consumidor_h4

# Limpiar
clean:
	@echo " Limpiando contenedores y volúmenes..."
	$(DOCKER_COMPOSE) down -v
	docker system prune -f
	@echo "✅ Limpieza completada"

# Recompilar Protocol Buffers
proto:
	@echo " Recompilando Protocol Buffers..."
	cd Broker_C1 && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
	cd Riploy_BD1_C2 && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
	cd Falabellox_BD2_C3 && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
	cd Parisio_BD3 && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
	cd Consumidores && protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
	@echo "✅ Protocol Buffers recompilados"

# Docker por VM (según especificación del laboratorio)
docker-VM1:
	@echo "🖥️  VM1: Riploy / BD1 / Consumidor2"
	$(DOCKER_COMPOSE) up -d riploy db1 consumidor_e2

docker-VM2:
	@echo "🖥️  VM2: Falabellox / BD2 / Consumidor3"
	$(DOCKER_COMPOSE) up -d falabellox db2 consumidor_e3

docker-VM3:
	@echo "🖥️  VM3: Parisio / BD3"
	$(DOCKER_COMPOSE) up -d parisio db3

docker-VM4:
	@echo "🖥️  VM4: Broker / Consumidor1"
	$(DOCKER_COMPOSE) up -d broker consumidor_e1

# Ver reporte generado
reporte:
	@echo " Contenido del Reporte.txt:"
	@docker exec cyberday_broker cat /root/Reporte.txt || echo "⚠️  Reporte aún no generado"

# Estado del sistema
status:
	@echo " Estado de servicios:"
	$(DOCKER_COMPOSE) ps

# Restart individual
restart-broker:
	$(DOCKER_COMPOSE) restart broker

restart-db:
	$(DOCKER_COMPOSE) restart db1 db2 db3

restart-prod:
	$(DOCKER_COMPOSE) restart riploy falabellox parisio

restart-cons:
	$(DOCKER_COMPOSE) restart consumidor_e1 consumidor_e2 consumidor_e3 consumidor_e4 \
		consumidor_m1 consumidor_m2 consumidor_m3 consumidor_m4 \
		consumidor_h1 consumidor_h2 consumidor_h3 consumidor_h4

# Tests
test:
	@echo " Ejecutando tests..."
	@for mod in Broker_C1 Riploy_BD1_C2 Falabellox_BD2_C3 Parisio_BD3 Consumidores; do \
		echo "--- Probando $$mod ---"; \
		(cd $$mod && go test ./... -v); \
	done
	@echo "✅ Pruebas completadas"

# Ver archivos CSV de consumidores
ver-consumidores:
	@echo " Archivos CSV de consumidores:"
	@docker exec cyberday_consumidor_e1 ls -lh /data/*.csv 2>/dev/null || echo "No hay archivos aún"

# Extraer resultados
extraer-resultados:
	@echo " Extrayendo resultados..."
	@mkdir -p resultados
	@docker cp cyberday_broker:/root/Reporte.txt ./resultados/ 2>/dev/null || echo "Reporte no disponible"
	@for i in e1 e2 e3 e4 m1 m2 m3 m4 h1 h2 h3 h4; do \
		docker cp cyberday_consumidor_$i:/data/C-${i^^}.csv ./resultados/ 2>/dev/null || true; \
	done
	@echo "✅ Resultados extraídos a ./resultados/"

# Monitoreo en tiempo real
monitor:
	@echo "📡 Monitoreo en tiempo real (Ctrl+C para salir)"
	watch -n 2 'docker-compose ps && echo "\n=== Logs recientes ===" && docker-compose logs --tail=5'