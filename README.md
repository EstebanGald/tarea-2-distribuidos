# CyberDay Distribuido - Laboratorio 2

Sistema distribuido de procesamiento de ofertas en tiempo real con tolerancia a fallos, simulando un evento CyberDay.

## Integrantes

- Vicente Luongo 202073637-5
- Antonio Rey 202173633-6
- Esteban Carrasco 201773546-5

## Descripción del Sistema

Sistema distribuido que simula el procesamiento de ofertas durante un CyberDay, implementando:

- **3 Productores** (Riploy, Falabellox, Parisio): Generan ofertas desde catálogos
- **1 Broker Central**: Valida, almacena y distribuye ofertas
- **3 Nodos de Base de Datos**: Almacenamiento distribuido con replicación (N=3, W=2, R=2)
- **12 Consumidores** (4 grupos × 3): Reciben ofertas según preferencias

### Características Principales

 **Tolerancia a fallos**: Simulación de caídas de nodos y consumidores  
 **Replicación eventual**: Modelo DynamoDB (N=3, W=2, R=2)  
 **Idempotencia**: Control de ofertas duplicadas  
 **Filtrado inteligente**: Distribución basada en preferencias  
 **Resincronización automática**: Recuperación tras fallos  
 **Persistencia**: Almacenamiento en disco de ofertas  

##  Arquitectura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Riploy    │     │  Falabellox │     │   Parisio   │
│ (Productor) │     │ (Productor) │     │ (Productor) │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       └───────────────────┼───────────────────┘
                           │ gRPC
                    ┌──────▼──────┐
                    │   BROKER    │
                    │  CENTRAL    │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐        ┌────▼────┐       ┌────▼────┐
   │   DB1   │◄──────►│   DB2   │◄─────►│   DB3   │
   │ (N=3)   │ Sync   │  (W=2)  │ Sync  │  (R=2)  │
   └─────────┘        └─────────┘       └─────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐        ┌────▼────┐       ┌────▼────┐
   │ Cons E1 │        │ Cons M1 │       │ Cons H1 │
   │ Cons E2 │        │ Cons M2 │       │ Cons H2 │
   │ Cons E3 │        │ Cons M3 │       │ Cons H3 │
   │ Cons E4 │        │ Cons M4 │       │ Cons H4 │
   └─────────┘        └─────────┘       └─────────┘
```

## Inicio Rápido

### Prerequisitos

- Docker >= 20.10
- Docker Compose >= 2.0
- Go >= 1.21 (solo para desarrollo)
- Make (opcional, para usar el Makefile)

### Instalación y Ejecución

```bash
# 1. Clonar el repositorio
git clone [URL_DEL_REPO]
cd CyberDay-Distribuido

# 2. Construir las imágenes
make build
# O sin Make:
docker-compose build

# 3. Iniciar el sistema completo
make up
# O sin Make:
docker-compose up -d

# 4. Ver logs en tiempo real
make logs
# O sin Make:
docker-compose logs -f

# 5. Ver estado de los servicios
make status
# O sin Make:
docker-compose ps
```

### Ejecución por Máquina Virtual

Según la especificación del laboratorio:

```bash
# MV1: Riploy / BD1 / Consumidor2
make docker-VM1

# MV2: Falabellox / BD2 / Consumidor3
make docker-VM2

# MV3: Parisio / BD3
make docker-VM3

# MV4: Broker / Consumidor1
make docker-VM4
```

## Monitoreo y Resultados

### Ver Logs por Componente

```bash
# Logs del broker
make logs-broker

# Logs de bases de datos
make logs-db

# Logs de productores
make logs-prod

# Logs de consumidores
make logs-cons
```

##  Estructura del Proyecto

```
.
├── Broker_C1/
│   ├── Broker/
│   │   └── broker_main.go      # Broker central completo
│   ├── proto/
│   │   └── ofertas.proto       # Definición Protocol Buffers
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── Riploy_BD1_C2/
│   ├── BD1/
│   │   ├── bd1.go              # Nodo de BD con replicación
│   │   └── Dockerfile
│   ├── Riploy/
│   │   ├── riploy.go           # Productor Riploy
│   │   ├── riploy_catalogo.csv
│   │   └── Dockerfile
│   ├── proto/
│   ├── go.mod
│   └── go.sum
│
├── Falabellox_BD2_C3/          # Estructura similar
├── Parisio_BD3/                # Estructura similar
│
├── Consumidores/
│   ├── consumidor.go           # Consumidor genérico
│   ├── proto/
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── consumidores.csv            # Preferencias de consumidores
├── docker-compose.yml          # Orquestación completa
├── Makefile                    # Comandos útiles
└── README.md
```

### Instrucciones para correr el programa

    En las máquinas virtuales, entrar a carpeta de tarea "tarea-2-distribuidos"
    Luego dependiendo de la máquina virtual, correr 'sudo make docker-VMn' n siendo el número de la VM

Para cada maquina virtual ejecutar:


make build
make docker-VM1 //en caso de ser la vm 1
make logs //para ver los logs de cada vm


VM1 (dist078) Ejecuta los contenedores de Riploy/BD1/Consumidor2

VM2 (dist079) Ejecuta los contenedores de Falabellox/BD2/Consumidor3

VM3 (dist080) Ejecuta los contenedores de Parisio/BD3

VM4 (dist102) Ejecuta los contenedores de Broker/Consumidor1