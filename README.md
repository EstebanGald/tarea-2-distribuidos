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

### Extraer Resultados

```bash
# Extraer reporte y archivos CSV
make extraer-resultados

# Los archivos se guardarán en ./resultados/
# - Reporte.txt: Resumen completo de la ejecución
# - C-E1.csv, C-E2.csv, ...: Ofertas recibidas por cada consumidor
```

### Ver Reporte

```bash
# Ver reporte directamente
make reporte

# O manualmente:
docker exec cyberday_broker cat /root/Reporte.txt
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

## Configuración

### Archivo consumidores.csv

Formato: `id_consumidor,categoria,tienda,precio_max`

```csv
C-E1,Electrónica,null,null
C-E2,Electrónica,Falabellox,null
C-E3,Electrónica,Riploy;Falabellox,null
C-E4,Electrónica,Parisio,100000
C-H3,Hogar;Electrónica;Deportes,null,300000
C-H4,null,null,null
```

**Reglas:**
- `null` = cualquier valor
- `;` = separador para múltiples valores
- Categorías y tiendas soportan múltiples valores
- `precio_max = 0` o `null` = sin límite

### Categorías Válidas

- Electrónica
- Moda
- Hogar
- Deportes
- Belleza
- Infantil
- Computación
- Electrodomésticos
- Herramientas
- Juguetes
- Automotriz
- Mascotas

## Simulación de Fallos

El sistema simula fallos automáticamente:

- **DB2**: Falla temporalmente a los 20 segundos por 15 segundos
- **Consumidor C-E3**: Se desconecta a los 30 segundos por 20 segundos
- **Consumidor C-H2**: Se desconecta a los 40 segundos por 15 segundos

### Comportamiento Esperado

1. **Durante fallo de DB2**:
   - Sistema continúa operando con DB1 y DB3
   - W=2 se cumple con los nodos activos
   - Ofertas se almacenan correctamente

2. **Recuperación de DB2**:
   - DB2 solicita sincronización de peers
   - Recibe ofertas perdidas durante el fallo
   - Se reintegra al sistema

3. **Durante desconexión de consumidor**:
   - Broker marca consumidor como inactivo
   - No se envían ofertas durante desconexión

4. **Recuperación de consumidor**:
   - Consumidor solicita histórico al broker
   - Broker consulta R=2 nodos DB
   - Consumidor recibe ofertas perdidas

## 📈 Parámetros de Consistencia

### N=3, W=2, R=2 (Modelo DynamoDB)

- **N=3**: Replicación en 3 nodos
- **W=2**: Escritura exitosa con 2 confirmaciones
- **R=2**: Lectura válida con 2 respuestas coincidentes

**Garantías:**
- Alta disponibilidad ante fallo de 1 nodo
- Consistencia eventual
- Durabilidad de datos

## Reporte Final

El broker genera `Reporte.txt` con:

### 1. Resumen de Productores
- Ofertas enviadas por productor
- Ofertas aceptadas
- Ofertas rechazadas

### 2. Estado de Nodos DB
- Estado final (activo/caído)
- Escrituras exitosas
- Escrituras fallidas

### 3. Notificaciones a Consumidores
- Ofertas recibidas por consumidor
- Estado final (activo/desconectado)
- Confirmación de archivos CSV

### 4. Fallos y Recuperaciones
- Listado de fallos simulados
- Resultados de resincronización

### 5. Conclusión
- Disponibilidad del sistema
- Cumplimiento de N=3, W=2, R=2

## 🛠️ Desarrollo

### Recompilar Protocol Buffers

```bash
make proto
```

O manualmente en cada módulo:

```bash
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
```

### Ejecutar Tests

```bash
make test
```

### Limpiar Sistema

```bash
# Detener y eliminar contenedores y volúmenes
make clean
```

## Troubleshooting

### Problema: Contenedores no se comunican

**Solución**: Verificar que todos los contenedores estén en la misma red

```bash
docker network inspect cyberday-distribuido_cyberday_network
```

### Problema: Broker no se conecta a nodos DB

**Solución**: Verificar que los nodos DB iniciaron primero

```bash
docker-compose logs db1 db2 db3
```

Reiniciar si es necesario:

```bash
make restart-db
make restart-broker
```

### Problema: Consumidores no reciben ofertas

**Solución**: Verificar registro en broker

```bash
docker-compose logs broker | grep "Registrando consumidor"
```

### Problema: Reporte no se genera

**Solución**: Esperar a que los productores terminen

```bash
docker-compose logs riploy falabellox parisio | grep "procesadas"
```

## Documentación Adicional

### Protocol Buffers

Ver `proto/ofertas.proto` para la definición completa de mensajes y servicios gRPC.

### Variables de Entorno

Cada componente acepta variables de entorno para configuración:

**Broker:**
- `TZ`: Zona horaria (default: America/Santiago)

**Nodos DB:**
- `NODO_ID`: Identificador del nodo (DB1, DB2, DB3)
- `PUERTO`: Puerto de escucha
- `TZ`: Zona horaria

**Productores:**
- `PRODUCTOR_NOMBRE`: Nombre del productor
- `CATALOGO`: Archivo CSV del catálogo
- `TZ`: Zona horaria

**Consumidores:**
- `CONSUMIDOR_ID`: ID del consumidor (C-E1, C-M2, etc.)
- `BROKER_ADDR`: Dirección del broker
- `ARCHIVO_CONFIG`: Ruta al consumidores.csv
- `TZ`: Zona horaria


---
