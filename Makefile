.PHONY: help build up down logs clean proto test

# ==========================================================
#         CONFIGURACIÓN DE IPs PARA MÚLTIPLES VMs
# ==========================================================
# Edita estas IPs con las direcciones reales de tus VMs
VM1_IP ?= 10.35.168.112 #dist102 <- CAMBIAR IP VM1 (Ej: Riploy, DB1, Consumidor E2)
VM2_IP ?= 10.35.168.88 #dist078 <- CAMBIAR IP VM2 (Ej: Falabellox, DB2, Consumidor E3)
VM3_IP ?= 10.35.168.89 #dist079 <- CAMBIAR IP VM3 (Ej: Parisio, DB3)
VM4_IP ?= 10.35.168.90 #dist080 <- CAMBIAR IP VM4 dist080(Ej: Broker, Consumidor E1)
# ==========================================================

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

# Docker por VM (CORREGIDO para usar nombres de imagen de docker-compose)

# --- VM1: Riploy / BD1 / Consumidor E2 ---
docker-VM1:
	@echo "🖥️  Iniciando servicios para VM1 (IP: $(VM1_IP))..."
	# Iniciar DB1
	@docker run -d --rm --name cyberday_db1 \
	  -v db1_data:/data \
	  -e TZ=America/Santiago \
	  -e NODO_ID=DB1 \
	  -e PUERTO=:50052 \
	  -e PEERS="$(VM2_IP):50053,$(VM3_IP):50054" \
	  --network=host \
	  tarea-2-distribuidos_db1 # <-- Nombre de imagen CORREGIDO
	@echo " -> DB1 iniciado."
	# Iniciar Riploy
	@docker run -d --rm --name cyberday_riploy \
	  -e TZ=America/Santiago \
	  -e PRODUCTOR_NOMBRE=Riploy \
	  -e CATALOGO=riploy_catalogo.csv \
	  -e BROKER_ADDR="$(VM4_IP):50051" \
	  --network=host \
	  tarea-2-distribuidos_riploy # <-- Nombre de imagen CORREGIDO
	@echo " -> Riploy iniciado."
	# Iniciar Consumidor E2
	@docker run -d --rm --name cyberday_consumidor_e2 \
	  -v /home/user/consumidores.csv:/app/consumidores.csv:ro \
	  -v consumidor_e2_data:/data \
	  -e TZ=America/Santiago \
	  -e CONSUMIDOR_ID=C-E2 \
	  -e BROKER_ADDR="$(VM4_IP):50051" \
	  --network=host \
	  tarea-2-distribuidos_consumidor_e2 # <-- Nombre de imagen CORREGIDO (ajustar si tienes nombres específicos por consumidor en docker-compose.yml)
	@echo " -> Consumidor E2 iniciado."
	@echo "✅ Servicios VM1 listos."

# --- VM2: Falabellox / BD2 / Consumidor E3 ---
docker-VM2:
	@echo "🖥️  Iniciando servicios para VM2 (IP: $(VM2_IP))..."
	# Iniciar DB2
	@docker run -d --rm --name cyberday_db2 \
	  -v db2_data:/data \
	  -e TZ=America/Santiago \
	  -e NODO_ID=DB2 \
	  -e PUERTO=:50053 \
	  -e PEERS="$(VM1_IP):50052,$(VM3_IP):50054" \
	  --network=host \
	  tarea-2-distribuidos_db2 # <-- Nombre de imagen CORREGIDO
	@echo " -> DB2 iniciado."
	# Iniciar Falabellox
	@docker run -d --rm --name cyberday_falabellox \
	  -e TZ=America/Santiago \
	  -e PRODUCTOR_NOMBRE=Falabellox \
	  -e CATALOGO=falabellox_catalogo.csv \
	  -e BROKER_ADDR="$(VM4_IP):50051" \
	  --network=host \
	  tarea-2-distribuidos_falabellox # <-- Nombre de imagen CORREGIDO
	@echo " -> Falabellox iniciado."
	# Iniciar Consumidor E3
	@docker run -d --rm --name cyberday_consumidor_e3 \
	  -v /home/user/consumidores.csv:/app/consumidores.csv:ro \
	  -v consumidor_e3_data:/data \
	  -e TZ=America/Santiago \
	  -e CONSUMIDOR_ID=C-E3 \
	  -e BROKER_ADDR="$(VM4_IP):50051" \
	  --network=host \
	  tarea-2-distribuidos_consumidor_e3 # <-- Nombre de imagen CORREGIDO (ajustar)
	@echo " -> Consumidor E3 iniciado."
	@echo "✅ Servicios VM2 listos."

# --- VM3: Parisio / BD3 ---
docker-VM3:
	@echo "🖥️  Iniciando servicios para VM3 (IP: $(VM3_IP))..."
	# Iniciar DB3
	@docker run -d --rm --name cyberday_db3 \
	  -v db3_data:/data \
	  -e TZ=America/Santiago \
	  -e NODO_ID=DB3 \
	  -e PUERTO=:50054 \
	  -e PEERS="$(VM1_IP):50052,$(VM2_IP):50053" \
	  --network=host \
	  tarea-2-distribuidos_db3 # <-- Nombre de imagen CORREGIDO
	@echo " -> DB3 iniciado."
	# Iniciar Parisio
	@docker run -d --rm --name cyberday_parisio \
	  -e TZ=America/Santiago \
	  -e PRODUCTOR_NOMBRE=Parisio \
	  -e CATALOGO=parisio_catalogo.csv \
	  -e BROKER_ADDR="$(VM4_IP):50051" \
	  --network=host \
	  tarea-2-distribuidos_parisio # <-- Nombre de imagen CORREGIDO
	@echo " -> Parisio iniciado."
	@echo "✅ Servicios VM3 listos."

# --- VM4: Broker / Consumidor E1 ---
docker-VM4:
	@echo "🖥️  Iniciando servicios para VM4 (IP: $(VM4_IP))..."
	# Iniciar Broker
	@docker run -d --rm --name cyberday_broker \
	  -e TZ=America/Santiago \
	  -e DB1_ADDR="$(VM1_IP):50052" \
	  -e DB2_ADDR="$(VM2_IP):50053" \
	  -e DB3_ADDR="$(VM3_IP):50054" \
	  --network=host \
	  tarea-2-distribuidos_broker # <-- Nombre de imagen CORREGIDO
	@echo " -> Broker iniciado."
	# Iniciar Consumidor E1
	@docker run -d --rm --name cyberday_consumidor_e1 \
	  -v /home/user/consumidores.csv:/app/consumidores.csv:ro \
	  -v consumidor_e1_data:/data \
	  -e TZ=America/Santiago \
	  -e CONSUMIDOR_ID=C-E1 \
	  -e BROKER_ADDR="127.0.0.1:50051" \
	  --network=host \
	  tarea-2-distribuidos_consumidor_e1 # <-- Nombre de imagen CORREGIDO (ajustar)
	@echo " -> Consumidor E1 iniciado."
	@echo "✅ Servicios VM4 listos."

# Comando para detener los contenedores en la VM actual
stop-vm-containers:
	@echo "🛑 Deteniendo contenedores Docker de CyberDay en esta VM..."
	@docker ps -q --filter "name=cyberday_" | xargs -r docker stop | xargs -r echo "Contenedores detenidos:"

# ... (resto del Makefile como logs, clean, proto, reporte, status, etc.)

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
# Prueba v3
monitor:
	@echo "📡 Monitoreo en tiempo real (Ctrl+C para salir)"
	watch -n 2 'docker-compose ps && echo "\n=== Logs recientes ===" && docker-compose logs --tail=5'