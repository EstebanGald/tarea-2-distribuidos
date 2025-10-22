package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings" 
	"time"

	pb "riploy_bd1_c2/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


var validCategorias = map[string]bool{
	"Electrónica":        true,
	"Moda":               true,
	"Hogar":              true,
	"Deportes":           true,
	"Belleza":            true,
	"Infantil":           true,
	"Computación":        true,
	"Electrodomésticos":  true,
	"Herramientas":       true,
	"Juguetes":           true,
	"Automotriz":         true,
	"Mascotas":           true,
}

type Productor struct {
	nombre    string
	catalogo  string
	client    pb.OfertasClient
	rand      *rand.Rand
}

func NewProductor(nombre, catalogo string) *Productor {
	return &Productor{
		nombre:   nombre,
		catalogo: catalogo,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Productor) conectarBroker(brokerAddr string) error {
	conn, err := grpc.Dial(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	
	p.client = pb.NewOfertasClient(conn)
	log.Printf("[%s] ✅ Conectado al broker", p.nombre)
	return nil
}

func (p *Productor) generarUUID() string {
	// Generar UUID simple: timestamp + random
	return fmt.Sprintf("%s-%d-%d", p.nombre, time.Now().UnixNano(), p.rand.Intn(999999))
}

func (p *Productor) validarYEnviarOferta(record []string) error {
	// record: [producto_id, tienda, categoria, producto, precio_base, stock]
	
	// Validar categoría
	categoria := record[2]
	if !validCategorias[categoria] {
		log.Printf("[%s] ⚠️  Categoría '%s' no válida, saltando", p.nombre, categoria)
		return fmt.Errorf("categoría no válida")
	}
	
	// Parsear precio y stock
	originalPrecioBase, err := strconv.Atoi(record[4])
	if err != nil {
		log.Printf("[%s] ⚠️  No se pudo transformar precio_base '%s', saltando", p.nombre, record[4])
		return err
	}
	
	stock, err := strconv.Atoi(record[5])
	if err != nil {
		log.Printf("[%s] ⚠️  No se pudo transformar stock '%s', saltando", p.nombre, record[5])
		return err
	}
	
	// Validar stock > 0
	if stock <= 0 {
		log.Printf("[%s] ⚠️  Stock = 0 para producto %s, saltando", p.nombre, record[0])
		return fmt.Errorf("stock inválido")
	}
	
	// Aplicar descuento aleatorio entre 10% y 50%
	discountPercent := 0.10 + p.rand.Float64()*0.40
	originalPrecioFloat := float64(originalPrecioBase)
	discountedPrecioFloat := originalPrecioFloat * (1.0 - discountPercent)
	finalPrecio := int32(discountedPrecioFloat)
	
	// Fecha actual
	currentTime := time.Now()
	formattedDate := currentTime.Format("2006-01-02")
	
	// Generar oferta_id único
	ofertaID := p.generarUUID()
	
	// Crear y enviar oferta
	oferta := &pb.OfertaRequest{
		OfertaId:        ofertaID,
		ProductoId:      record[0],
		Tienda:          record[1],
		Categoria:       categoria,
		Producto:        record[3],
		PrecioDescuento: finalPrecio,
		Stock:           int32(stock),
		Fecha:           formattedDate,
		ClienteId:       p.nombre,
		Timestamp:       time.Now().Unix(),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	resp, err := p.client.EnviarOferta(ctx, oferta)
	if err != nil {
		log.Printf("[%s] ❌ Error enviando oferta %s: %v", p.nombre, record[0], err)
		return err
	}
	
	if resp.GetExito() {
		log.Printf("[%s] ✅ Oferta %s enviada: %s - $%d (desc: %.0f%%)", 
			p.nombre, record[0], record[3], finalPrecio, discountPercent*100)
	} else {
		log.Printf("[%s] ⚠️  Oferta %s rechazada: %s", p.nombre, record[0], resp.GetMensaje())
	}
	
	return nil
}

func (p *Productor) procesarCatalogo() error {
	file, err := os.Open(p.catalogo)
	if err != nil {
		return err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	
	// Saltar header
	if _, err := reader.Read(); err != nil {
		return err
	}
	
	ofertasEnviadas := 0
	ofertasExitosas := 0
	ofertasRechazadas := 0
	
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[%s] Error leyendo línea del CSV: %v", p.nombre, err)
			continue
		}
		
		ofertasEnviadas++
		
		// Validar y enviar
		if err := p.validarYEnviarOferta(record); err != nil {
			ofertasRechazadas++
		} else {
			ofertasExitosas++
		}
		
		// Esperar tiempo aleatorio entre 500ms y 2000ms
		sleepDuration := time.Duration(500+p.rand.Intn(1500)) * time.Millisecond
		time.Sleep(sleepDuration)
	}
	
	log.Printf("[%s] 📊 RESUMEN:", p.nombre)
	log.Printf("  - Total intentadas: %d", ofertasEnviadas)
	log.Printf("  - Exitosas: %d", ofertasExitosas)
	log.Printf("  - Rechazadas: %d", ofertasRechazadas)
	
	return nil
}

func main() {
	// Leer configuración desde variables de entorno
	nombre := os.Getenv("PRODUCTOR_NOMBRE")
	if nombre == "" {
		if len(os.Args) > 1 {
			nombre = os.Args[1]
		} else {
			nombre = "Riploy" // Default
		}
	}
	
	catalogo := os.Getenv("CATALOGO")
	if catalogo == "" {
		catalogo = fmt.Sprintf("%s_catalogo.csv", strings.ToLower(nombre))
	}

	brokerAddr := os.Getenv("BROKER_ADDR")
    if brokerAddr == "" {
        brokerAddr = "broker:50051" // Mantenemos un default
    }
	
	log.Printf("[PRODUCTOR] Iniciando %s", nombre)
	log.Printf("[PRODUCTOR] Catálogo: %s", catalogo)
	
	productor := NewProductor(nombre, catalogo)
	
	// Conectar al broker con reintentos
	var err error // <-- Declare err *before* the loop
	for intentos := 0; intentos < 10; intentos++ {
		err := productor.conectarBroker(brokerAddr)
		if err == nil {
			break
		}
		log.Printf("[%s] Error conectando (intento %d/10): %v", nombre, intentos+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil { // Check the error from the loop
		log.Fatalf("[%s] No se pudo conectar al broker después de 10 intentos: %v", nombre, err)
	}
	
	if productor.client == nil {
		log.Fatalf("[%s] No se pudo conectar al broker después de 10 intentos", nombre)
	}
	
	// Esperar un poco antes de empezar a enviar
	time.Sleep(5 * time.Second)
	
	// Procesar catálogo y enviar ofertas
	if err := productor.procesarCatalogo(); err != nil {
		log.Fatalf("[%s] Error procesando catálogo: %v", nombre, err)
	}
	
	log.Printf("[%s] ✅ Todas las ofertas del catálogo han sido procesadas", nombre)
}