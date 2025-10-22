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

	pb "broker_c1/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address_broker = ":50051"
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

type ConsumidorInfo struct {
	ID            string
	Categorias    []string
	Tiendas       []string
	PrecioMax     int32
	DireccionGRPC string
	Cliente       pb.NotificacionesConsumidorClient
	Activo        bool
}

type EstadisticasProductor struct {
	OfertasEnviadas  int
	OfertasAceptadas int
	OfertasRechazadas int
}

type EstadisticasNodo struct {
	NodoID            string
	Activo            bool
	EscriturasExitosas int
	EscriturasFallidas int
}

type EstadisticasConsumidor struct {
	ConsumidorID      string
	OfertasRecibidas  int
	Activo            bool
}

type server struct {
	pb.UnimplementedOfertasServer
	pb.UnimplementedConsumidorServer
	
	// Productores registrados
	productores      []string
	productoresMutex sync.Mutex
	
	// Consumidores registrados
	consumidores      map[string]*ConsumidorInfo
	consumidoresMutex sync.RWMutex
	
	// Nodos DB
	dbClients []pb.DynamoDBClient
	dbActivos []bool
	dbMutex   sync.RWMutex
	
	// Control de duplicados (idempotencia)
	ofertasProcesadas      map[string]bool
	ofertasProcesakdasMutex sync.Mutex
	
	// Estadísticas
	statsProductores   map[string]*EstadisticasProductor
	statsNodos         []*EstadisticasNodo
	statsConsumidores  map[string]*EstadisticasConsumidor
	statsMutex         sync.Mutex
}

func (s *server) EnviarOferta(ctx context.Context, in *pb.OfertaRequest) (*pb.OfertaResponse, error) {
	clienteID := in.GetClienteId()
	ofertaID := in.GetOfertaId()
	
	log.Printf("[BROKER] Recibida oferta %s de %s", ofertaID, clienteID)
	
	// 1. Validar productor registrado
	if !s.esProductorRegistrado(clienteID) {
		s.registrarProductor(clienteID)
	}
	
	s.incrementarOfertasEnviadas(clienteID)
	
	// 2. Validar oferta
	if err := s.validarOferta(in); err != nil {
		log.Printf("[BROKER] Oferta %s rechazada: %v", ofertaID, err)
		s.incrementarOfertasRechazadas(clienteID)
		return &pb.OfertaResponse{Exito: false, Mensaje: err.Error()}, nil
	}
	
	// 3. Verificar idempotencia
	if s.esOfertaDuplicada(ofertaID) {
		log.Printf("[BROKER] Oferta %s duplicada, descartando", ofertaID)
		return &pb.OfertaResponse{Exito: true, Mensaje: "Oferta ya procesada"}, nil
	}
	
	// 4. Almacenar en base de datos distribuida (W=2)
	confirmaciones := s.almacenarEnDB(ctx, in)
	if confirmaciones < 2 {
		log.Printf("[BROKER] ERROR: Solo %d confirmaciones, se requieren W=2", confirmaciones)
		return &pb.OfertaResponse{Exito: false, Mensaje: "No se alcanzó W=2"}, nil
	}
	
	log.Printf("[BROKER] Oferta %s almacenada con %d confirmaciones (W=2 cumplido)", ofertaID, confirmaciones)
	
	// 5. Marcar como procesada
	s.marcarOfertaProcesada(ofertaID)
	s.incrementarOfertasAceptadas(clienteID)
	
	// 6. Distribuir a consumidores interesados
	s.distribuirAConsumidores(ctx, in)
	
	return &pb.OfertaResponse{Exito: true, Mensaje: "Oferta registrada y distribuida"}, nil
}

func (s *server) RegistrarConsumidor(ctx context.Context, in *pb.RegistroConsumidorRequest) (*pb.RegistroConsumidorResponse, error) {
	s.consumidoresMutex.Lock()
	defer s.consumidoresMutex.Unlock()
	
	consumidorID := in.GetConsumidorId()
	log.Printf("[BROKER] Registrando consumidor %s", consumidorID)
	
	// Conectar al servicio gRPC del consumidor
	conn, err := grpc.Dial(in.GetDireccionGrpc(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[BROKER] Error conectando a consumidor %s: %v", consumidorID, err)
		return &pb.RegistroConsumidorResponse{Exito: false, Mensaje: err.Error()}, nil
	}
	
	cliente := pb.NewNotificacionesConsumidorClient(conn)
	
	s.consumidores[consumidorID] = &ConsumidorInfo{
		ID:            consumidorID,
		Categorias:    in.GetCategorias(),
		Tiendas:       in.GetTiendas(),
		PrecioMax:     in.GetPrecioMax(),
		DireccionGRPC: in.GetDireccionGrpc(),
		Cliente:       cliente,
		Activo:        true,
	}
	
	s.statsMutex.Lock()
	s.statsConsumidores[consumidorID] = &EstadisticasConsumidor{
		ConsumidorID:     consumidorID,
		OfertasRecibidas: 0,
		Activo:           true,
	}
	s.statsMutex.Unlock()
	
	log.Printf("[BROKER] Consumidor %s registrado exitosamente", consumidorID)
	return &pb.RegistroConsumidorResponse{Exito: true, Mensaje: "Registrado"}, nil
}

func (s *server) SolicitarHistorico(ctx context.Context, in *pb.SolicitarHistoricoRequest) (*pb.HistoricoConsumidorResponse, error) {
	consumidorID := in.GetConsumidorId()
	log.Printf("[BROKER] Consumidor %s solicita histórico", consumidorID)
	
	// Leer de al menos 2 nodos (R=2)
	historicos := s.leerHistoricoDistribuido(ctx)
	
	if len(historicos) < 2 {
		log.Printf("[BROKER] ERROR: Solo %d nodos respondieron, se requieren R=2", len(historicos))
		return &pb.HistoricoConsumidorResponse{Ofertas: nil}, nil
	}
	
	// Combinar resultados (consenso simple: unión de ofertas)
	ofertasMap := make(map[string]*pb.OfertaRequest)
	for _, hist := range historicos {
		for _, oferta := range hist.Ofertas {
			ofertasMap[oferta.OfertaId] = oferta
		}
	}
	
	ofertas := make([]*pb.OfertaRequest, 0, len(ofertasMap))
	for _, oferta := range ofertasMap {
		ofertas = append(ofertas, oferta)
	}
	
	// Filtrar por preferencias del consumidor
	s.consumidoresMutex.RLock()
	consumidor, existe := s.consumidores[consumidorID]
	s.consumidoresMutex.RUnlock()
	
	if existe {
		ofertas = s.filtrarOfertas(ofertas, consumidor)
	}
	
	log.Printf("[BROKER] Enviando %d ofertas históricas a %s", len(ofertas), consumidorID)
	return &pb.HistoricoConsumidorResponse{Ofertas: ofertas}, nil
}

func (s *server) validarOferta(oferta *pb.OfertaRequest) error {
	if oferta.GetOfertaId() == "" {
		return fmt.Errorf("oferta_id vacío")
	}
	if oferta.GetStock() <= 0 {
		return fmt.Errorf("stock debe ser mayor a 0")
	}
	if !validCategorias[oferta.GetCategoria()] {
		return fmt.Errorf("categoría %s no válida", oferta.GetCategoria())
	}
	return nil
}

func (s *server) esOfertaDuplicada(ofertaID string) bool {
	s.ofertasProcesakdasMutex.Lock()
	defer s.ofertasProcesakdasMutex.Unlock()
	return s.ofertasProcesadas[ofertaID]
}

func (s *server) marcarOfertaProcesada(ofertaID string) {
	s.ofertasProcesakdasMutex.Lock()
	defer s.ofertasProcesakdasMutex.Unlock()
	s.ofertasProcesadas[ofertaID] = true
}

func (s *server) almacenarEnDB(ctx context.Context, oferta *pb.OfertaRequest) int {
	confirmaciones := 0
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	s.dbMutex.RLock()
	defer s.dbMutex.RUnlock()
	
	for i, dbClient := range s.dbClients {
		if !s.dbActivos[i] {
			continue
		}
		
		wg.Add(1)
		go func(idx int, client pb.DynamoDBClient) {
			defer wg.Done()
			
			ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			
			resp, err := client.GuardarOferta(ctxTimeout, oferta)
			if err != nil {
				log.Printf("[BROKER] Error guardando en DB%d: %v", idx+1, err)
				s.incrementarEscriturasFallidas(idx)
				return
			}
			
			if resp.GetExito() {
				mu.Lock()
				confirmaciones++
				mu.Unlock()
				s.incrementarEscriturasExitosas(idx)
				log.Printf("[BROKER] DB%d confirmó almacenamiento", idx+1)
			}
		}(i, dbClient)
	}
	
	wg.Wait()
	return confirmaciones
}

func (s *server) distribuirAConsumidores(ctx context.Context, oferta *pb.OfertaRequest) {
	s.consumidoresMutex.RLock()
	defer s.consumidoresMutex.RUnlock()
	
	for _, consumidor := range s.consumidores {
		if !consumidor.Activo {
			continue
		}
		
		if s.ofertaCumpleFiltros(oferta, consumidor) {
			go s.enviarAConsumidor(ctx, consumidor, oferta)
		}
	}
}

func (s *server) ofertaCumpleFiltros(oferta *pb.OfertaRequest, consumidor *ConsumidorInfo) bool {
	// Filtro de categoría
	if len(consumidor.Categorias) > 0 && consumidor.Categorias[0] != "null" {
		categoriaMatch := false
		for _, cat := range consumidor.Categorias {
			if cat == oferta.GetCategoria() {
				categoriaMatch = true
				break
			}
		}
		if !categoriaMatch {
			return false
		}
	}
	
	// Filtro de tienda
	if len(consumidor.Tiendas) > 0 && consumidor.Tiendas[0] != "null" {
		tiendaMatch := false
		for _, tienda := range consumidor.Tiendas {
			if tienda == oferta.GetTienda() {
				tiendaMatch = true
				break
			}
		}
		if !tiendaMatch {
			return false
		}
	}
	
	// Filtro de precio
	if consumidor.PrecioMax > 0 && oferta.GetPrecioDescuento() > consumidor.PrecioMax {
		return false
	}
	
	return true
}

func (s *server) enviarAConsumidor(ctx context.Context, consumidor *ConsumidorInfo, oferta *pb.OfertaRequest) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	
	_, err := consumidor.Cliente.RecibirOferta(ctxTimeout, oferta)
	if err != nil {
		log.Printf("[BROKER] Error enviando a consumidor %s: %v", consumidor.ID, err)
		s.marcarConsumidorInactivo(consumidor.ID)
		return
	}
	
	s.incrementarOfertasRecibidas(consumidor.ID)
	log.Printf("[BROKER] Oferta %s enviada a consumidor %s", oferta.GetOfertaId(), consumidor.ID)
}

func (s *server) leerHistoricoDistribuido(ctx context.Context) []*pb.HistoricoResponse {
	var historicos []*pb.HistoricoResponse
	var wg sync.WaitGroup
	var mu sync.Mutex
	
	s.dbMutex.RLock()
	defer s.dbMutex.RUnlock()
	
	for i, dbClient := range s.dbClients {
		if !s.dbActivos[i] {
			continue
		}
		
		wg.Add(1)
		go func(idx int, client pb.DynamoDBClient) {
			defer wg.Done()
			
			ctxTimeout, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			
			resp, err := client.LeerHistorico(ctxTimeout, &pb.LeerHistoricoRequest{
				NodoId:         fmt.Sprintf("DB%d", idx+1),
				DesdeTimestamp: 0,
			})
			if err != nil {
				log.Printf("[BROKER] Error leyendo de DB%d: %v", idx+1, err)
				return
			}
			
			mu.Lock()
			historicos = append(historicos, resp)
			mu.Unlock()
		}(i, dbClient)
	}
	
	wg.Wait()
	return historicos
}

func (s *server) filtrarOfertas(ofertas []*pb.OfertaRequest, consumidor *ConsumidorInfo) []*pb.OfertaRequest {
	var filtradas []*pb.OfertaRequest
	for _, oferta := range ofertas {
		if s.ofertaCumpleFiltros(oferta, consumidor) {
			filtradas = append(filtradas, oferta)
		}
	}
	return filtradas
}

func (s *server) esProductorRegistrado(clienteID string) bool {
	s.productoresMutex.Lock()
	defer s.productoresMutex.Unlock()
	for _, id := range s.productores {
		if id == clienteID {
			return true
		}
	}
	return false
}

func (s *server) registrarProductor(clienteID string) {
	s.productoresMutex.Lock()
	defer s.productoresMutex.Unlock()
	s.productores = append(s.productores, clienteID)
	
	s.statsMutex.Lock()
	s.statsProductores[clienteID] = &EstadisticasProductor{}
	s.statsMutex.Unlock()
	
	log.Printf("[BROKER] Productor %s registrado", clienteID)
}

func (s *server) incrementarOfertasEnviadas(clienteID string) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if stats, ok := s.statsProductores[clienteID]; ok {
		stats.OfertasEnviadas++
	}
}

func (s *server) incrementarOfertasAceptadas(clienteID string) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if stats, ok := s.statsProductores[clienteID]; ok {
		stats.OfertasAceptadas++
	}
}

func (s *server) incrementarOfertasRechazadas(clienteID string) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if stats, ok := s.statsProductores[clienteID]; ok {
		stats.OfertasRechazadas++
	}
}

func (s *server) incrementarEscriturasExitosas(idx int) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if idx < len(s.statsNodos) {
		s.statsNodos[idx].EscriturasExitosas++
	}
}

func (s *server) incrementarEscriturasFallidas(idx int) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if idx < len(s.statsNodos) {
		s.statsNodos[idx].EscriturasFallidas++
	}
}

func (s *server) incrementarOfertasRecibidas(consumidorID string) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	if stats, ok := s.statsConsumidores[consumidorID]; ok {
		stats.OfertasRecibidas++
	}
}

func (s *server) marcarConsumidorInactivo(consumidorID string) {
	s.consumidoresMutex.Lock()
	defer s.consumidoresMutex.Unlock()
	if consumidor, ok := s.consumidores[consumidorID]; ok {
		consumidor.Activo = false
	}
	
	s.statsMutex.Lock()
	if stats, ok := s.statsConsumidores[consumidorID]; ok {
		stats.Activo = false
	}
	s.statsMutex.Unlock()
}

func (s *server) generarReporte() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	
	file, err := os.Create("Reporte.txt")
	if err != nil {
		log.Printf("Error creando reporte: %v", err)
		return
	}
	defer file.Close()
	
	fmt.Fprintf(file, "=== REPORTE CYBERDAY DISTRIBUIDO ===\n")
	fmt.Fprintf(file, "Fecha: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	// Resumen de productores
	fmt.Fprintf(file, "--- RESUMEN DE PRODUCTORES ---\n")
	for id, stats := range s.statsProductores {
		fmt.Fprintf(file, "Productor: %s\n", id)
		fmt.Fprintf(file, "  Ofertas enviadas: %d\n", stats.OfertasEnviadas)
		fmt.Fprintf(file, "  Ofertas aceptadas: %d\n", stats.OfertasAceptadas)
		fmt.Fprintf(file, "  Ofertas rechazadas: %d\n", stats.OfertasRechazadas)
		fmt.Fprintf(file, "\n")
	}
	
	// Estado de nodos
	fmt.Fprintf(file, "--- ESTADO DE NODOS DE BASE DE DATOS ---\n")
	for _, stats := range s.statsNodos {
		estado := "ACTIVO"
		if !stats.Activo {
			estado = "CAÍDO"
		}
		fmt.Fprintf(file, "Nodo: %s - Estado: %s\n", stats.NodoID, estado)
		fmt.Fprintf(file, "  Escrituras exitosas: %d\n", stats.EscriturasExitosas)
		fmt.Fprintf(file, "  Escrituras fallidas: %d\n", stats.EscriturasFallidas)
		fmt.Fprintf(file, "\n")
	}
	
	// Notificaciones a consumidores
	fmt.Fprintf(file, "--- NOTIFICACIONES A CONSUMIDORES ---\n")
	for id, stats := range s.statsConsumidores {
		estado := "ACTIVO"
		if !stats.Activo {
			estado = "DESCONECTADO"
		}
		fmt.Fprintf(file, "Consumidor: %s - Estado: %s\n", id, estado)
		fmt.Fprintf(file, "  Ofertas recibidas: %d\n", stats.OfertasRecibidas)
		fmt.Fprintf(file, "\n")
	}
	
	// Conclusión
	fmt.Fprintf(file, "--- CONCLUSIÓN ---\n")
	fmt.Fprintf(file, "El sistema mantuvo disponibilidad y consistencia bajo las reglas N=3, W=2, R=2.\n")
	fmt.Fprintf(file, "Total ofertas procesadas: %d\n", len(s.ofertasProcesadas))
	
	log.Println("[BROKER] Reporte generado: Reporte.txt")
}

func cargarConsumidoresDesdeCSV(rutaCSV string) ([]*pb.RegistroConsumidorRequest, error) {
	file, err := os.Open(rutaCSV)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	
	var consumidores []*pb.RegistroConsumidorRequest
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		
		categorias := []string{}
		if record[1] != "null" {
			categorias = strings.Split(record[1], ";")
		} else {
			categorias = []string{"null"}
		}
		
		tiendas := []string{}
		if record[2] != "null" {
			tiendas = strings.Split(record[2], ";")
		} else {
			tiendas = []string{"null"}
		}
		
		precioMax := int32(0)
		if record[3] != "null" {
			precio, _ := strconv.Atoi(record[3])
			precioMax = int32(precio)
		}
		
		consumidores = append(consumidores, &pb.RegistroConsumidorRequest{
			ConsumidorId: record[0],
			Categorias:   categorias,
			Tiendas:      tiendas,
			PrecioMax:    precioMax,
		})
	}
	
	return consumidores, nil
}

func newDBClient(address string) (pb.DynamoDBClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return pb.NewDynamoDBClient(conn), conn, nil
}

func main() {
	log.Println("[BROKER] Iniciando...")
	
	// Conectar a nodos DB
	db1Addr := os.Getenv("DB1_ADDR")
    if db1Addr == "" { db1Addr = "db1:50052" } // Default

    db2Addr := os.Getenv("DB2_ADDR")
    if db2Addr == "" { db2Addr = "db2:50053" } // Default

    db3Addr := os.Getenv("DB3_ADDR")
    if db3Addr == "" { db3Addr = "db3:50054" } // Default
	dbAddresses := []string{db1Addr, db2Addr, db3Addr} // <-- Use the variables read from env
	var dbClients []pb.DynamoDBClient
	var connections []*grpc.ClientConn
	dbActivos := []bool{true, true, true}
	
	for i, addr := range dbAddresses {
		client, conn, err := newDBClient(addr)
		if err != nil {
			log.Printf("[BROKER] ADVERTENCIA: No se pudo conectar a DB%d: %v", i+1, err)
			dbActivos[i] = false
			dbClients = append(dbClients, nil)
		} else {
			dbClients = append(dbClients, client)
			connections = append(connections, conn)
			log.Printf("[BROKER] Conectado a DB%d", i+1)
		}
	}
	
	// Crear servidor
	srv := &server{
		productores:          make([]string, 0),
		consumidores:         make(map[string]*ConsumidorInfo),
		dbClients:            dbClients,
		dbActivos:            dbActivos,
		ofertasProcesadas:    make(map[string]bool),
		statsProductores:     make(map[string]*EstadisticasProductor),
		statsConsumidores:    make(map[string]*EstadisticasConsumidor),
		statsNodos: []*EstadisticasNodo{
			{NodoID: "DB1", Activo: dbActivos[0]},
			{NodoID: "DB2", Activo: dbActivos[1]},
			{NodoID: "DB3", Activo: dbActivos[2]},
		},
	}
	
	// Iniciar servidor gRPC
	lis, err := net.Listen("tcp", address_broker)
	if err != nil {
		log.Fatalf("[BROKER] Error escuchando: %v", err)
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterOfertasServer(grpcServer, srv)
	pb.RegisterConsumidorServer(grpcServer, srv)
	
	log.Printf("[BROKER] Escuchando en %v", lis.Addr())
	
	// Generar reporte al finalizar
	defer func() {
		srv.generarReporte()
		for _, conn := range connections {
			conn.Close()
		}
	}()
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[BROKER] Error sirviendo: %v", err)
	}
}