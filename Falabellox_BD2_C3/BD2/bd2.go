package main

import (
	"context"
	"log"
	"net"
	"sync"

	// Importamos el código generado por protoc
	pb "falabellox_bd2_c3/proto"

	"google.golang.org/grpc"
)

const (
	adress_bd2 = ":50053" // Usar un puerto diferente al del broker
)

type offerStore struct {
	offers map[string]*pb.OfertasRequest
	mu     sync.Mutex // Mutex to protect the map from concurrent writes
}

// Definimos una struct para nuestro servidor. Debe embeber el UnimplementedGreeterServer.
// Esto asegura la compatibilidad hacia adelante si se añaden más RPCs al servicio.
type server struct {
	pb.UnimplementedDynamoDBServer
	store *offerStore
}

func (s *server) GuardarOfertas(ctx context.Context, in *pb.OfertasRequest) (*pb.OfertasResponse, error) {
	log.Printf("Recibida petición de Broker para almacenar oferta de %v", in.GetClienteId())
	clientID := in.GetClienteId()
	s.store.mu.Lock()
	// Guardamos la oferta en el mapa
	s.store.offers[clientID] = in
	//Print store.offers content
	log.Printf("Oferta almacenada para el cliente ID: %s", s.store.offers[clientID])
	// Unlock el mutex
	s.store.mu.Unlock()
	return &pb.OfertasResponse{BrokerMessage: "ACK"}, nil
}

func main() {

	lis, err := net.Listen("tcp", adress_bd2)
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	store := &offerStore{
		offers: make(map[string]*pb.OfertasRequest),
	}

	// 2. Creamos una nueva instancia del servidor gRPC
	s := grpc.NewServer()
	pb.RegisterDynamoDBServer(s, &server{store: store})
	log.Printf("Servidor bd2 escuchando en %v", lis.Addr())

	// 4. Iniciamos el servidor para que empiece a aceptar peticiones en el puerto.
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
