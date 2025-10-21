package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "Parisio_BD3/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type DBNode struct {
	pb.UnimplementedDynamoDBServer
	
	nodoID          string
	puerto          string
	ofertas         map[string]*pb.OfertaRequest
	ofertasMutex    sync.RWMutex
	
	peers           []string
	peerClients     []pb.DynamoDBClient
	peersMutex      sync.RWMutex
	
	activo          bool
	estadoMutex     sync.RWMutex
	
	archivoPersistencia string
}

func NewDBNode(nodoID, puerto string, peers []string) *DBNode {
	return &DBNode{
		nodoID:              nodoID,
		puerto:              puerto,
		ofertas:             make(map[string]*pb.OfertaRequest),
		peers:               peers,
		peerClients:         make([]pb.DynamoDBClient, len(peers)),
		activo:              true,
		archivoPersistencia: fmt.Sprintf("%s_ofertas.json", nodoID),
	}
}

func (db *DBNode) GuardarOferta(ctx context.Context, in *pb.OfertaRequest) (*pb.AckResponse, error) {
	db.estadoMutex.RLock()
	activo := db.activo
	db.estadoMutex.RUnlock()
	
	if !activo {
		return &pb.AckResponse{
			Exito:   false,
			NodoId:  db.nodoID,
			Mensaje: "Nodo inactivo",
		}, nil
	}
	
	ofertaID := in.GetOfertaId()
	log.Printf("[%s] Guardando oferta %s", db.nodoID, ofertaID)
	
	db.ofertasMutex.Lock()
	db.ofertas[ofertaID] = in
	db.ofertasMutex.Unlock()
	
	if err := db.persistirOfertas(); err != nil {
		log.Printf("[%s] Error persistiendo: %v", db.nodoID, err)
	}
	
	return &pb.AckResponse{
		Exito:   true,
		NodoId:  db.nodoID,
		Mensaje: "ACK",
	}, nil
}

func (db *DBNode) LeerHistorico(ctx context.Context, in *pb.LeerHistoricoRequest) (*pb.HistoricoResponse, error) {
	log.Printf("[%s] Leyendo histórico", db.nodoID)
	
	db.ofertasMutex.RLock()
	defer db.ofertasMutex.RUnlock()
	
	ofertas := make([]*pb.OfertaRequest, 0, len(db.ofertas))
	
	for _, oferta := range db.ofertas {
		if in.GetDesdeTimestamp() > 0 && oferta.GetTimestamp() < in.GetDesdeTimestamp() {
			continue
		}
		ofertas = append(ofertas, oferta)
	}
	
	log.Printf("[%s] Devolviendo %d ofertas", db.nodoID, len(ofertas))
	
	return &pb.HistoricoResponse{
		Ofertas: ofertas,
		NodoId:  db.nodoID,
	}, nil
}

func (db *DBNode) Sincronizar(ctx context.Context, in *pb.SincronizarRequest) (*pb.SincronizarResponse, error) {
	log.Printf("[%s] Recibiendo sincronización de %s", db.nodoID, in.GetNodoOrigen())
	
	ofertasSincronizadas := 0
	
	db.ofertasMutex.Lock()
	for _, oferta := range in.GetOfertas() {
		ofertaID := oferta.GetOfertaId()
		if _, existe := db.ofertas[ofertaID]; !existe {
			db.ofertas[ofertaID] = oferta
			ofertasSincronizadas++
		}
	}
	db.ofertasMutex.Unlock()
	
	if ofertasSincronizadas > 0 {
		db.persistirOfertas()
	}
	
	log.Printf("[%s] Sincronizadas %d ofertas nuevas", db.nodoID, ofertasSincronizadas)
	
	return &pb.SincronizarResponse{
		Exito:                true,
		OfertasSincronizadas: int32(ofertasSincronizadas),
	}, nil
}

func (db *DBNode) persistirOfertas() error {
	db.ofertasMutex.RLock()
	defer db.ofertasMutex.RUnlock()
	
	file, err := os.Create(db.archivoPersistencia)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	return encoder.Encode(db.ofertas)
}

func (db *DBNode) cargarOfertas() error {
	file, err := os.Open(db.archivoPersistencia)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[%s] No hay archivo de persistencia previo", db.nodoID)
			return nil
		}
		return err
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	
	db.ofertasMutex.Lock()
	defer db.ofertasMutex.Unlock()
	
	err = decoder.Decode(&db.ofertas)
	if err != nil && err != io.EOF {
		return err
	}
	
	log.Printf("[%s] Cargadas %d ofertas desde disco", db.nodoID, len(db.ofertas))
	return nil
}

func (db *DBNode) conectarAPeers() {
	db.peersMutex.Lock()
	defer db.peersMutex.Unlock()
	
	for i, peerAddr := range db.peers {
		conn, err := grpc.Dial(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("[%s] Error conectando a peer %s: %v", db.nodoID, peerAddr, err)
			continue
		}
		
		db.peerClients[i] = pb.NewDynamoDBClient(conn)
		log.Printf("[%s] Conectado a peer %s", db.nodoID, peerAddr)
	}
}

func (db *DBNode) sincronizarConPeers() {
	db.ofertasMutex.RLock()
	ofertas := make([]*pb.OfertaRequest, 0, len(db.ofertas))
	for _, oferta := range db.ofertas {
		ofertas = append(ofertas, oferta)
	}
	db.ofertasMutex.RUnlock()
	
	if len(ofertas) == 0 {
		return
	}
	
	db.peersMutex.RLock()
	defer db.peersMutex.RUnlock()
	
	for i, peerClient := range db.peerClients {
		if peerClient == nil {
			continue
		}
		
		go func(idx int, client pb.DynamoDBClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			resp, err := client.Sincronizar(ctx, &pb.SincronizarRequest{
				NodoOrigen: db.nodoID,
				Ofertas:    ofertas,
			})
			
			if err != nil {
				log.Printf("[%s] Error sincronizando con peer %d: %v", db.nodoID, idx, err)
				return
			}
			
			log.Printf("[%s] Peer %d sincronizó %d ofertas", db.nodoID, idx, resp.GetOfertasSincronizadas())
		}(i, peerClient)
	}
}

func (db *DBNode) simularFallo(duracion time.Duration) {
	log.Printf("[%s] ⚠️  SIMULANDO FALLO POR %v", db.nodoID, duracion)
	
	db.estadoMutex.Lock()
	db.activo = false
	db.estadoMutex.Unlock()
	
	time.Sleep(duracion)
	
	db.estadoMutex.Lock()
	db.activo = true
	db.estadoMutex.Unlock()
	
	log.Printf("[%s] ✅ RECUPERADO DE FALLO - Iniciando resincronización", db.nodoID)
	
	time.Sleep(2 * time.Second)
	db.solicitarSincronizacionDePeers()
}

func (db *DBNode) solicitarSincronizacionDePeers() {
	db.peersMutex.RLock()
	defer db.peersMutex.RUnlock()
	
	for i, peerClient := range db.peerClients {
		if peerClient == nil {
			continue
		}
		
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		resp, err := peerClient.LeerHistorico(ctx, &pb.LeerHistoricoRequest{
			NodoId:         db.nodoID,
			DesdeTimestamp: 0,
		})
		cancel()
		
		if err != nil {
			log.Printf("[%s] Error solicitando histórico de peer %d: %v", db.nodoID, i, err)
			continue
		}
		
		db.ofertasMutex.Lock()
		nuevasOfertas := 0
		for _, oferta := range resp.GetOfertas() {
			if _, existe := db.ofertas[oferta.GetOfertaId()]; !existe {
				db.ofertas[oferta.GetOfertaId()] = oferta
				nuevasOfertas++
			}
		}
		db.ofertasMutex.Unlock()
		
		if nuevasOfertas > 0 {
			db.persistirOfertas()
			log.Printf("[%s] Resincronizadas %d ofertas desde peer %d", db.nodoID, nuevasOfertas, i)
		}
		
		break
	}
}

func main() {
	nodoID := os.Getenv("NODO_ID")
	if nodoID == "" {
		nodoID = "DB3"  // ← DEFAULT PARA BD3
	}
	
	puerto := os.Getenv("PUERTO")
	if puerto == "" {
		puerto = ":50054"  // ← PUERTO DEFAULT PARA BD3
	}
	
	peers := []string{}
	if nodoID == "DB1" {
		peers = []string{"db2:50053", "db3:50054"}
	} else if nodoID == "DB2" {
		peers = []string{"db1:50052", "db3:50054"}
	} else if nodoID == "DB3" {
		peers = []string{"db1:50052", "db2:50053"}
	}
	
	dbNode := NewDBNode(nodoID, puerto, peers)
	
	if err := dbNode.cargarOfertas(); err != nil {
		log.Printf("[%s] Error cargando ofertas: %v", nodoID, err)
	}
	
	go func() {
		time.Sleep(3 * time.Second)
		dbNode.conectarAPeers()
		
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			dbNode.sincronizarConPeers()
		}
	}()
	
	// ← BD1 NO SIMULA FALLO (comentado o eliminado)
	// NO incluir el bloque de simulación de fallo
	
	lis, err := net.Listen("tcp", puerto)
	if err != nil {
		log.Fatalf("[%s] Error escuchando: %v", nodoID, err)
	}
	
	grpcServer := grpc.NewServer()
	pb.RegisterDynamoDBServer(grpcServer, dbNode)
	
	log.Printf("[%s] Escuchando en %v", nodoID, lis.Addr())
	
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[%s] Error sirviendo: %v", nodoID, err)
	}
}