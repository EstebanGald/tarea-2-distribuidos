# Resumen de Dockerfiles - CyberDay Distribuido

## 📁 Ubicación de Cada Dockerfile

```
CyberDay-Distribuido/
│
├── Broker_C1/
│   └── Dockerfile                          ← Broker Central
│
├── Riploy_BD1_C2/
│   ├── BD1/
│   │   └── Dockerfile                      ← Nodo DB1
│   └── Riploy/
│       └── Dockerfile                      ← Productor Riploy
│
├── Falabellox_BD2_C3/
│   ├── BD2/
│   │   └── Dockerfile                      ← Nodo DB2
│   └── Falabellox/
│       └── Dockerfile                      ← Productor Falabellox
│
├── Parisio_BD3/
│   ├── BD3/
│   │   └── Dockerfile                      ← Nodo DB3
│   └── Parisio/
│       └── Dockerfile                      ← Productor Parisio
│
└── Consumidores/
    └── Dockerfile                          ← Consumidores (genérico)
```

## 🔧 Características de Cada Dockerfile

### 1. Broker Central (`Broker_C1/Dockerfile`)

**Características:**
- Build multi-stage para optimizar tamaño
- Usuario no-root para seguridad
- Healthcheck en puerto 50051
- Timezone configurado
- Expone puerto 50051

**Comando de construcción:**
```bash
cd Broker_C1
docker build -t cyberday-broker .
```

**Tamaño aproximado:** ~15 MB

---

### 2. Nodos DB (`BD1/Dockerfile`, `BD2/Dockerfile`, `BD3/Dockerfile`)

**Características:**
- Build multi-stage
- Usuario no-root
- Healthcheck en puertos 50052/50053/50054
- Volumen en `/data` para persistencia
- netcat-openbsd instalado para healthchecks
- Variables de entorno: `NODO_ID`, `PUERTO`

**Comando de construcción:**
```bash
# DB1
cd Riploy_BD1_C2/BD1
docker build -t cyberday-db1 .

# DB2
cd Falabellox_BD2_C3/BD2
docker build -t cyberday-db2 .

# DB3
cd Parisio_BD3/BD3
docker build -t cyberday-db3 .
```

**Tamaño aproximado:** ~12 MB cada uno

---

### 3. Productores (`Riploy/Dockerfile`, `Falabellox/Dockerfile`, `Parisio/Dockerfile`)

**Características:**
- Build multi-stage
- Usuario no-root
- Incluye catálogo CSV dentro de la imagen
- Variables de entorno: `PRODUCTOR_NOMBRE`, `CATALOGO`
- No requiere volúmenes externos

**Comando de construcción:**
```bash
# Riploy
cd Riploy_BD1_C2/Riploy
docker build -t cyberday-riploy .

# Falabellox
cd Falabellox_BD2_C3/Falabellox
docker build -t cyberday-falabellox .

# Parisio
cd Parisio_BD3/Parisio
docker build -t cyberday-parisio .
```

**Tamaño aproximado:** ~10 MB cada uno

---

### 4. Consumidores (`Consumidores/Dockerfile`)

**Características:**
- Build multi-stage
- Usuario no-root
- Healthcheck flexible (múltiples puertos)
- Volumen en `/data` para guardar CSVs
- `consumidores.csv` montado desde host
- Variables de entorno: `CONSUMIDOR_ID`, `BROKER_ADDR`, `ARCHIVO_CONFIG`
- Expone puertos 50061-50072

**Comando de construcción:**
```bash
cd Consumidores
docker build -t cyberday-consumidor .
```

**Tamaño aproximado:** ~12 MB

**Nota:** Esta imagen se reutiliza para los 12 consumidores con diferentes variables de entorno.

---

## 🚀 Construcción de Todas las Imágenes

### Opción 1: Usando Docker Compose (Recomendado)

```bash
# En el directorio raíz del proyecto
docker-compose build
```

Esto construirá automáticamente todas las imágenes definidas en `docker-compose.yml`.

### Opción 2: Usando Makefile

```bash
make build
```

### Opción 3: Manualmente

```bash
# Script para construir todas las imágenes
#!/bin/bash

echo "Construyendo Broker..."
cd Broker_C1 && docker build -t cyberday-broker . && cd ..

echo "Construyendo DB1..."
cd Riploy_BD1_C2/BD1 && docker build -t cyberday-db1 . && cd ../..

echo "Construyendo DB2..."
cd Falabellox_BD2_C3/BD2 && docker build -t cyberday-db2 . && cd ../..

echo "Construyendo DB3..."
cd Parisio_BD3/BD3 && docker build -t cyberday-db3 . && cd ../..

echo "Construyendo Riploy..."
cd Riploy_BD1_C2/Riploy && docker build -t cyberday-riploy . && cd ../..

echo "Construyendo Falabellox..."
cd Falabellox_BD2_C3/Falabellox && docker build -t cyberday-falabellox . && cd ../..

echo "Construyendo Parisio..."
cd Parisio_BD3/Parisio && docker build -t cyberday-parisio . && cd ../..

echo "Construyendo Consumidores..."
cd Consumidores && docker build -t cyberday-consumidor . && cd ..

echo "✅ Todas las imágenes construidas"
```

---

## 📊 Comparación de Imágenes

| Componente | Imagen Base | Binario | Extras | Tamaño Total |
|------------|-------------|---------|--------|--------------|
| Broker | alpine:latest | Go binary | CA certs, tzdata | ~15 MB |
| DB1/2/3 | alpine:latest | Go binary | CA certs, tzdata, netcat | ~12 MB |
| Productores | alpine:latest | Go binary + CSV | CA certs, tzdata | ~10 MB |
| Consumidores | alpine:latest | Go binary | CA certs, tzdata, netcat | ~12 MB |

**Total aproximado:** ~100 MB para todas las imágenes

---

## 🔍 Verificación de Imágenes

Después de construir, verifica que todas las imágenes existan:

```bash
docker images | grep cyberday
```

Deberías ver:

```
cyberday-broker         latest
cyberday-db1           latest
cyberday-db2           latest
cyberday-db3           latest
cyberday-riploy        latest
cyberday-falabellox    latest
cyberday-parisio       latest
cyberday-consumidor    latest
```

---

## 🛠️ Troubleshooting

### Error: "go.mod not found"

**Solución:** Asegúrate de que cada directorio tenga su `go.mod` y `go.sum`.

```bash
# En cada directorio con código Go
go mod init [nombre-modulo]
go mod tidy
```

### Error: "COPY failed: file not found"

**Solución:** Verifica la estructura de directorios. Los Dockerfiles asumen:

```
Riploy_BD1_C2/
├── go.mod          ← En la raíz del módulo
├── go.sum
├── proto/          ← Proto compartido
├── BD1/
│   ├── bd1.go
│   └── Dockerfile  ← Copia ../go.mod
└── Riploy/
    ├── riploy.go
    ├── riploy_catalogo.csv
    └── Dockerfile  ← Copia ../go.mod
```

### Error: "cannot find package"

**Solución:** Regenera los archivos proto:

```bash
cd [directorio]
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
```

### Imágenes demasiado grandes

**Solución:** Los Dockerfiles ya están optimizados con multi-stage builds, pero puedes:

```bash
# Limpiar cache de construcción
docker builder prune

# Ver capas de una imagen
docker history cyberday-broker
```

---

## 🎯 Mejores Prácticas Implementadas

✅ **Multi-stage builds** - Reducen tamaño final  
✅ **Usuarios no-root** - Mayor seguridad  
✅ **Healthchecks** - Monitoreo automático  
✅ **Variables de entorno** - Configuración flexible  
✅ **Imágenes Alpine** - Mínimo tamaño  
✅ **Cache de dependencias** - Builds más rápidos  
✅ **Permisos correctos** - Seguridad en runtime  

---

## 📝 Notas Adicionales

### Build Context

Cada Dockerfile usa su directorio actual como build context. Para DB y Productores, necesitan acceso al directorio padre para `go.mod` y `proto/`.

### Orden de Build

Docker Compose maneja automáticamente el orden, pero manualmente:

1. Primero: Broker y DBs (no dependen de nadie)
2. Después: Productores (dependen del Broker)
3. Finalmente: Consumidores (dependen del Broker)

### Cache de Docker

Para forzar rebuild sin cache:

```bash
docker-compose build --no-cache
# o
docker build --no-cache -t [nombre] .
```

---

## 🔗 Referencias

- [Docker Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Alpine Linux](https://www.alpinelinux.org/)

---

¡Todos los Dockerfiles están optimizados y listos para producción! 🐳