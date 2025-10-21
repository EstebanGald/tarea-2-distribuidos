package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "consumidor/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Consumidor struct {
	pb.UnimplementedNotificacionesConsumidorServer
	
	id            string
	categorias    []string
	tiendas       []string
	precioMax     int32
	puerto        string
	
	ofertas       []*pb.OfertaRequest
	ofertasMutex  sync.Mutex
	
	archivoCSV    string
	
	// Cliente para conectarse al broker
	brokerClient  pb.ConsumidorClient
	
	// Estado
	activo        bool
	estadoMutex   sync.RWMutex
}

func NewConsumidor(id string, categorias, tiendas []string, precioMax int32, puerto string) *Consumidor {
	return &Consumidor{
		id:         id,
		categorias: categorias,
		tiendas:    tiendas,
		precioMax:  precioMax,
		puerto:     puerto,
		ofertas:    make([]*pb.OfertaRequest, 0),
		archivoCSV: fmt.Sprintf("%s.csv", id),
		activo:     true,
	}
}

func (c *Consumidor) RecibirOferta(ctx context.Context, in *pb.OfertaRequest) (*pb.AckResponse, error) {
	c.estadoMutex.RLock()
	activo := c.activo
	c.estadoMutex.RUnlock()
	
	if !activo {
		return &pb.AckResponse{
			Exito:   false,
			Mensaje: "Consumidor inactivo",
		}, nil
	}
	
	log.Printf("[%s] 📦 Recibida oferta %s: %s - $%d", 
		c.id, in.GetOfertaId(), in.GetProducto(), in.GetPrecioDescuento())
	
	// Almacenar oferta
	c.ofertasMutex.Lock()
	c.ofertas = append(c.ofertas, in)
	c.ofertasMutex.Unlock()
	
	// Guardar en CSV
	if err := c.guardarEnCSV(in); err != nil {
		log.Printf("[%s] Error guardando en CSV: %v", c.id, err)
	}
	
	return &pb.AckResponse{
		Exito:   true,
		NodoId:  c.id,
		Mensaje: "Oferta recibida",
	}, nil
}

func (c *Consumidor) guardarEnCSV(oferta *pb.OfertaRequest) error {
	c.ofertasMutex.Lock()
	defer c.ofertasMutex.Unlock()
	
	// Verificar si el archivo existe
	fileExists := true
	if _, err := os.Stat(c.archivoCSV); os.IsNotExist(err) {
		fileExists = false
	}
	
	file, err := os.OpenFile(c.archivoCSV, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Escribir header si es archivo nuevo
	if !fileExists {
		header := []string{"oferta_id", "producto_id", "tienda", "categoria", "producto", "precio_descuento", "stock", "fecha", "timestamp"}
		if err := writer.Write(header); err != nil {
			return err
		}
	}
	
	// Escribir fila
	row := []string{
		oferta.GetOfertaId(),
		oferta.GetProductoId(),
		oferta.GetTienda(),
		oferta.GetCategoria(),
		oferta.GetProducto(),
		fmt.Sprintf("%d", oferta.GetPrecioDescuento()),
		fmt.Sprintf("%d", oferta.GetStock()),
		oferta.GetFecha(),
		fmt.Sprintf("%d", oferta.GetTimestamp()),
	}
	
	return writer.Write(row)
}

func (c *Consumidor) registrarEnBroker(brokerAddr string) error {
	conn, err := grpc.Dial(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	
	c.brokerClient = pb.NewConsumidorClient(conn)
	
	// Obtener IP del contenedor/host
	miDireccion := fmt.Sprintf("%s%s", c.id, c.puerto)
	
	resp, err := c.brokerClient.RegistrarConsumidor(context.Background(), &pb.RegistroConsumidorRequest{
		ConsumidorId:   c.id,
		Categorias:     c.categorias,
		Tiendas:        c.tiendas,
		PrecioMax:      c.precioMax,
		DireccionGrpc:  miDireccion,
	})
	
	if err != nil {
		return err
	}
	
	if !resp.GetExito() {
		return fmt.Errorf("broker rechazó registro: %s", resp.GetMensaje())
	}
	
	log.Printf("[%s] ✅ Registrado exitosamente en el broker", c.id)
	return nil
}

func (c *Consumidor) solicitarHistorico() error {
	if c.brokerClient == nil {
		return fmt.Errorf("no conectado al broker")
	}
	
	log.Printf("[%s] 🔍 Solicitando histórico al broker...", c.id)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	resp, err := c.brokerClient.SolicitarHistorico(ctx, &pb.SolicitarHistoricoRequest{
		ConsumidorId: c.id,
	})
	
	if err != nil {
		return err
	}
	
	log.Printf("[%s] 📚 Recibidas %d ofertas históricas", c.id, len(resp.GetOfertas()))
	
	// Guardar ofertas históricas
	for _, oferta := range resp.GetOfertas() {
		// Verificar si ya tenemos esta oferta
		existe := false
		c.ofertasMutex.Lock()
		for _, o := range c.ofertas {
			if o.GetOfertaId() == oferta.GetOfertaId() {
				existe = true
				break
			}
		}
		if !existe {
			c.ofertas = append(c.ofertas, oferta)
			c.guardarEnCSV(oferta)
		}
		c.ofertasMutex.Unlock()
	}
	
	return nil
}

func (c *Consumidor) simularDesconexion(duracion time.Duration) {
	log.Printf("[%s] ⚠️  SIMULANDO DESCONEXIÓN POR %v", c.id, duracion)
	
	c.estadoMutex.Lock()
	c.activo = false
	c.estadoMutex.Unlock()
	
	time.Sleep(duracion)
	
	c.estadoMutex.Lock()
	c.activo = true
	c.estadoMutex.Unlock()
	
	log.Printf("[%s] ✅ RECONECTADO - Solicitando histórico", c.id)
	
	// Esperar un poco para estabilizar
	time.Sleep(2 * time.Second)
	
	// Solicitar histórico perdido
	if err := c.solicitarHistorico(); err != nil {
		log.Printf("[%s] Error solicitando histórico: %v", c.id, err)
	}
}

func cargarPreferenciasConsumidor(archivoCSV, consumidorID string) (*Consumidor, error) {
	file, err := os.Open(archivoCSV)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		
		if record[0] == consumidorID {
			// Parsear categorías
			categorias := []string{}
			if record[1] != "null" {
				categorias = strings.Split(record[1], ";")
			} else {
				categorias = []string{"null"}
			}
			
			// Parsear tiendas
			tiendas := []string{}
			if record[2] != "null" {
				tiendas = strings.Split(record[2], ";")
			} else {
				tiendas = []string{"null"}
			}
			
			// Parsear precio_max
			precioMax := int32(0)
			if record[3] != "null" {
				precio, _ := strconv.Atoi(record[3])
				precioMax = int32(precio)
			}
			
			// Determinar puerto basado en ID
			puerto := determinarPuerto(consumidorID)
			
			return NewConsumidor(consumidorID, categorias, tiendas, precioMax, puerto), nil
		}
	}
	
	return nil, fmt.Errorf("consumidor %s no encontrado en CSV", consumidorID)
}

func determinarPuerto(consumidorID string) string {
	// Asignar puertos únicos basados en el ID
	puertos := map[string]string{
		"C-E1": ":50061",
		"C-E2": ":50062",
		"C-E3": ":50063",
		"C-E4": ":50064",
		"C-M1": ":50065",
		"C-M2": ":50066",
		"C-M3": ":50067",
		"C-M4": ":50068",
		"C-H1": ":50069",
		"C-H2": ":50070",
		"C-H3": ":50071",
		"C-H4": ":50072",
	}
	
	if puerto, existe := puertos[consumidorID]; existe {
		return puerto
	}
	
	return ":50061" // Default
}

func main() {
	// Leer ID del consumidor desde argumentos o variable de entorno
	consumidorID := os.Getenv("CONSUMIDOR_ID")
	if consumidorID == "" {
		if len(os.Args) > 1 {
			consumidorID = os.Args[1]
		} else {
			consumidorID = "C-E1" // Default
		}
	}
	
	archivoConfig := os.Getenv("ARCHIVO_CONFIG")
	if archivoConfig == "" {
		archivoConfig = "consumidores.csv"
	}
	
	brokerAddr := os.Getenv("BROKER_ADDR")
	if brokerAddr == "" {
		brokerAddr = "broker:50051"
	}
	
	log.Printf("[CONSUMIDOR] Iniciando consumidor %s", consumidorID)
	
	// Cargar preferencias desde CSV
	consumidor, err := cargarPreferenciasConsumidor(archivoConfig, consumidorID)
	if err != nil {
		log.Fatalf("Error cargando preferencias: %v", err)
	}
	
	log.Printf("[%s] Preferencias:", consumidor.id)
	log.Printf("  - Categorías: %v", consumidor.categorias)
	log.Printf("  - Tiendas: %v", consumidor.tiendas)
	log.Printf("  - Precio máximo: %d", consumidor.precioMax)
	
	// Iniciar servidor gRPC para recibir ofertas
	lis, err := net.Listen("tcp", consumidor.puerto)
	if err != nil {
		log.Fatalf("Error escuchando: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterNotificacionesConsumidorServer(grpcServer, consumidor)
	
	go func() {
		log.Printf("[%s] Escuchando en %v", consumidor.id, lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Error sirviendo: %v", err)
		}
	}()
	
	// Esperar un poco para que el servidor inicie
	time.Sleep(2 * time.Second)
	
	// Registrarse en el broker
	for intentos := 0; intentos < 5; intentos++ {
		err = consumidor.registrarEnBroker(brokerAddr)
		if err == nil {
			break
		}
		log.Printf("[%s] Error registrando (intento %d/5): %v", consumidor.id, intentos+1, err)
		time.Sleep(3 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("[%s] No se pudo registrar en el broker después de 5 intentos", consumidor.id)
	}
	
	// Simular desconexión para algunos consumidores
	if consumidorID == "C-E3" {
		go func() {
			time.Sleep(30 * time.Second)
			consumidor.simularDesconexion(20 * time.Second)
		}()
	} else if consumidorID == "C-H2" {
		go func() {
			time.Sleep(40 * time.Second)
			consumidor.simularDesconexion(15 * time.Second)
		}()
	}
	
	// Mantener el programa corriendo
	log.Printf("[%s] ✅ Consumidor activo y esperando ofertas...", consumidor.id)
	
	// Bloquear indefinidamente
	select {}
}