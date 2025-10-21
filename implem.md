# Guía de Implementación - CyberDay Distribuido

Esta guía detalla paso a paso cómo implementar el código proporcionado en tu repositorio existente.

## 📋 Checklist de Archivos a Crear/Modificar

### ✅ Archivos Nuevos a Crear

```
✓ consumidores.csv
✓ docker-compose.yml
✓ Makefile
✓ setup.sh
✓ test_system.sh
✓ .gitignore
✓ Consumidores/consumidor.go
✓ Consumidores/Dockerfile
✓ Consumidores/go.mod
✓ Consumidores/proto/ofertas.proto
```

### 🔄 Archivos a Modificar

```
✓ Broker_C1/Broker/broker_main.go (REEMPLAZAR COMPLETAMENTE)
✓ Broker_C1/proto/ofertas.proto (REEMPLAZAR COMPLETAMENTE)
✓ Riploy_BD1_C2/BD1/bd1.go (REEMPLAZAR COMPLETAMENTE)
✓ Riploy_BD1_C2/Riploy/riploy.go (REEMPLAZAR COMPLETAMENTE)
✓ Falabellox_BD2_C3/BD2/bd2.go (REEMPLAZAR COMPLETAMENTE)
✓ Falabellox_BD2_C3/Falabellox/falabellox.go (REEMPLAZAR COMPLETAMENTE)
✓ Parisio_BD3/BD3/bd3.go (REEMPLAZAR COMPLETAMENTE)
✓ Parisio_BD3/Parisio/parisio.go (REEMPLAZAR COMPLETAMENTE)
✓ README.md (REEMPLAZAR COMPLETAMENTE)
```

## 🚀 Pasos de Implementación

### Paso 1: Backup del Código Actual

```bash
# Crear backup de tu código actual
cd /ruta/a/tu/proyecto
git add .
git commit -m "Backup antes de implementar mejoras"
git tag backup-antes-mejoras
```

### Paso 2: Crear Estructura de Directorios

```bash
# Crear directorio para consumidores
mkdir -p Consumidores/proto

# Crear directorios de utilidad
mkdir -p resultados
mkdir -p shared
```

### Paso 3: Actualizar Protocol Buffers

**Archivo: `Broker_C1/proto/ofertas.proto`**

Reemplazar completamente con el contenido del artifact `ofertas_proto_mejorado`.

**Importante:** Este mismo archivo debe copiarse a:
- `Riploy_BD1_C2/proto/ofertas.proto`
- `Falabellox_BD2_C3/proto/ofertas.proto`
- `Parisio_BD3/proto/ofertas.proto`
- `Consumidores/proto/ofertas.proto`

```bash
# Copiar proto actualizado a todos los módulos
cp Broker_C1/proto/ofertas.proto Riploy_BD1_C2/proto/
cp Broker_C1/proto/ofertas.proto Falabellox_BD2_C3/proto/
cp Broker_C1/proto/ofertas.proto Parisio_BD3/proto/
cp Broker_C1/proto/ofertas.proto Consumidores/proto/
```

### Paso 4: Recompilar Protocol Buffers

```bash
# En cada directorio con proto/
cd Broker_C1
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto

cd ../Riploy_BD1_C2
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto

cd ../Falabellox_BD2_C3
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto

cd ../Parisio_BD3
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto

cd ../Consumidores
protoc --go_out=. --go-grpc_out=. proto/ofertas.proto
```

### Paso 5: Actualizar Código del Broker

**Archivo: `Broker_C1/Broker/broker_main.go`**

Reemplazar completamente con el contenido del artifact `broker_completo`.

### Paso 6: Actualizar Código de Nodos DB

**Para cada nodo (BD1, BD2, BD3):**

Reemplazar el archivo correspondiente:
- `Riploy_BD1_C2/BD1/bd1.go`
- `Falabellox_BD2_C3/BD2/bd2.go`
- `Parisio_BD3/BD3/bd3.go`

Con el contenido del artifact `nodo_db_completo`, ajustando las rutas de importación:

```go
// Para BD1
import pb "riploy_bd1_c2/proto"

// Para BD2
import pb "falabellox_bd2_c3/proto"

// Para BD3
import pb "parisio_bd3/proto"
```

### Paso 7: Actualizar Código de Productores

**Para cada productor:**

Reemplazar los archivos:
- `Riploy_BD1_C2/Riploy/riploy.go`
- `Falabellox_BD2_C3/Falabellox/falabellox.go`
- `Parisio_BD3/Parisio/parisio.go`

Con el contenido del artifact `productor_mejorado`, ajustando imports:

```go
// Para Riploy
import pb "riploy_bd1_c2/proto"

// Para Falabellox
import pb "falabellox_bd2_c3/proto"

// Para Parisio
import pb "parisio_bd3/proto"
```

**Importante:** Añadir línea faltante al inicio:

```go
import "strings" // Añadir esta importación
```

### Paso 8: Crear Módulo de Consumidores

**Archivo: `Consumidores/consumidor.go`**

Crear con el contenido del artifact `consumidor_completo`.

**Archivo: `Consumidores/go.mod`**

```go
module consumidor

go 1.24.2

require (
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
)
```

**Ejecutar en Consumidores/:**

```bash
cd Consumidores
go mod download
```

### Paso 9: Crear Dockerfiles

**Archivo: `Broker_C1/Dockerfile`**

Usar contenido del artifact `dockerfile_broker`.

**Archivo: `Riploy_BD1_C2/BD1/Dockerfile`** (y similares para BD2, BD3)

Usar contenido del artifact `dockerfile_db`, ajustando rutas.

**Archivo: `Riploy_BD1_C2/Riploy/Dockerfile`** (y similares para otros productores)

Usar contenido del artifact `dockerfile_productor`.

**Archivo: `Consumidores/Dockerfile`**

Usar contenido del artifact `dockerfile_consumidor`.

### Paso 10: Crear Archivos de Configuración

**Archivo: `consumidores.csv`** (en raíz del proyecto)

Usar contenido del artifact `consumidores_config`.

**Archivo: `docker-compose.yml`** (en raíz del proyecto)

Usar contenido del artifact `docker_compose`.

**Archivo: `Makefile`** (en raíz del proyecto)

Usar contenido del artifact `makefile`.

**Archivo: `.gitignore`** (en raíz del proyecto)

Usar contenido del artifact `gitignore`.

### Paso 11: Crear Scripts de Utilidad

**Archivo: `setup.sh`** (en raíz del proyecto)

Usar contenido del artifact `setup_script`.

```bash
chmod +x setup.sh
```

**Archivo: `test_system.sh`** (en raíz del proyecto)

Usar contenido del artifact `test_script`.

```bash
chmod +x test_system.sh
```

### Paso 12: Actualizar README

**Archivo: `README.md`** (en raíz del proyecto)

Reemplazar con el contenido del artifact `readme_completo`.

### Paso 13: Verificar Estructura Final

Tu proyecto debe tener esta estructura:

```
.
├── Broker_C1/
│   ├── Broker/
│   │   └── broker_main.go          ← ACTUALIZADO
│   ├── proto/
│   │   ├── ofertas.proto           ← ACTUALIZADO
│   │   ├── ofertas.pb.go           ← REGENERAR
│   │   └── ofertas_grpc.pb.go      ← REGENERAR
│   ├── Dockerfile                  ← NUEVO
│   ├── go.mod
│   └── go.sum
│
├── Riploy_BD1_C2/
│   ├── BD1/
│   │   ├── bd1.go                  ← ACTUALIZADO
│   │   └── Dockerfile              ← NUEVO
│   ├── Riploy/
│   │   ├── riploy.go               ← ACTUALIZADO
│   │   ├── riploy_catalogo.csv
│   │   └── Dockerfile              ← NUEVO
│   ├── proto/                      ← ACTUALIZADO
│   ├── go.mod
│   └── go.sum
│
├── Falabellox_BD2_C3/             ← Similar estructura
├── Parisio_BD3/                   ← Similar estructura
│
├── Consumidores/                  ← DIRECTORIO NUEVO
│   ├── consumidor.go              ← NUEVO
│   ├── proto/                     ← NUEVO
│   ├── Dockerfile                 ← NUEVO
│   ├── go.mod                     ← NUEVO
│   └── go.sum                     ← NUEVO
│
├── consumidores.csv               ← NUEVO
├── docker-compose.yml             ← NUEVO
├── Makefile                       ← NUEVO
├── setup.sh                       ← NUEVO
├── test_system.sh                 ← NUEVO
├── .gitignore                     ← NUEVO
└── README.md                      ← ACTUALIZADO
```

### Paso 14: Ejecutar Setup

```bash
# Hacer ejecutables los scripts
chmod +x setup.sh test_system.sh

# Ejecutar setup
./setup.sh
```

Esto verificará prerequisitos, creará directorios, recompilará proto y construirá las imágenes Docker.

### Paso 15: Probar el Sistema

```bash
# Iniciar sistema completo
make up

# Ver logs
make logs

# En otra terminal, después de 2 minutos
./test_system.sh
```

### Paso 16: Verificar Resultados

```bash
# Extraer resultados
make extraer-resultados

# Ver reporte
cat resultados/Reporte.txt

# Ver CSVs de consumidores
ls -lh resultados/*.csv
```

## 🔧 Solución de Problemas Comunes

### Problema: Errores de Compilación de Proto

**Solución:**

```bash
# Instalar protoc si no está instalado
# En Ubuntu/Debian:
sudo apt-get install protobuf-compiler

# En macOS:
brew install protobuf

# Instalar plugins de Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Asegurarse de que estén en PATH
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Problema: Errores de Dependencias de Go

**Solución:**

```bash
# En cada directorio con go.mod
go mod tidy
go mod download
```

### Problema: Contenedores No Se Comunican

**Solución:**

```bash
# Verificar red de Docker
docker network ls
docker network inspect cyberday-distribuido_cyberday_network

# Recrear red si es necesario
make down
docker network prune
make up
```

### Problema: "localhost" No Resuelve en Docker

**Acción Requerida:**

En todos los archivos `.go`, reemplazar direcciones `localhost` por nombres de servicio:

```go
// ANTES:
const address_broker = "localhost:50051"

// DESPUÉS:
const address_broker = "broker:50051"
```

Esto ya está hecho en los artifacts proporcionados.

## ✅ Checklist Final

Antes de la entrega, verificar:

- [ ] Todos los archivos `.proto` son idénticos
- [ ] Todos los archivos `.pb.go` están regenerados
- [ ] Todos los `go.mod` tienen las dependencias correctas
- [ ] Todos los Dockerfiles existen y son correctos
- [ ] `docker-compose.yml` está completo
- [ ] `consumidores.csv` existe en la raíz
- [ ] `Makefile` funciona correctamente
- [ ] `setup.sh` ejecuta sin errores
- [ ] `make build` construye todas las imágenes
- [ ] `make up` inicia todos los servicios
- [ ] `./test_system.sh` pasa todos los tests
- [ ] `make extraer-resultados` genera archivos
- [ ] `Reporte.txt` se genera correctamente
- [ ] Los 12 consumidores generan sus CSVs
- [ ] README.md está actualizado
- [ ] `.gitignore` está configurado
- [ ] Código está comentado y formateado

## 📝 Notas Adicionales

### Sobre los Imports

Cada módulo usa su propio package name en imports:

```go
// Broker_C1
import pb "broker_c1/proto"

// Riploy_BD1_C2
import pb "riploy_bd1_c2/proto"

// Falabellox_BD2_C3
import pb "falabellox_bd2_c3/proto"

// Parisio_BD3
import pb "parisio_bd3/proto"

// Consumidores
import pb "consumidor/proto"
```

### Sobre Direcciones de Red

En Docker, usar nombres de servicio (no `localhost`):
- `broker:50051`
- `db1:50052`
- `db2:50053`
- `db3:50054`

### Sobre los Puertos

Mapeo de puertos en docker-compose:
- Broker: 50051
- DB1: 50052
- DB2: 50053
- DB3: 50054
- Consumidores: 50061-50072

## 🎯 Próximos Pasos

1. Implementar el código siguiendo esta guía
2. Ejecutar `./setup.sh`
3. Probar con `make up`
4. Ejecutar tests con `./test_system.sh`
5. Verificar resultados
6. Hacer commit y push
7. Preparar entrega

¡Éxito con el laboratorio!