package main

import (
	"context"
	"log"
	"net"
	"sync"

	// Importamos el código generado por protoc
	pb "broker_main/proto" // Reemplaza con el path de tu módulo

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"  // 1. Import for gRPC error codes
	"google.golang.org/grpc/status" // 2. Import for gRPC error status
)

// Definimos una struct para nuestro servidor. Debe embeber el UnimplementedGreeterServer.
// Esto asegura la compatibilidad hacia adelante si se añaden más RPCs al servicio.
type server struct {
	pb.UnimplementedOfertasServer
	registeredClients []string // Lista para almacenar los IDs de clientes registrados
	clientMutex       sync.Mutex
}

func (s *server) Ofertas(ctx context.Context, in *pb.OfertasRequest) (*pb.OfertasResponse, error) {
	clientID := in.GetClienteId()
	//log.Printf("Recibida petición de: %v", in.GetClienteId())
	if clientID == "" {
		log.Println("Rejecting request with empty client ID.")
		return nil, status.Errorf(codes.InvalidArgument, "Client ID is required.")
	}
	s.clientMutex.Lock()
	// Verificamos si el cliente ya está registrado
	isRegistered := false
	for _, id := range s.registeredClients {
		if id == clientID {
			isRegistered = true
			break
		}
	}
	if !isRegistered {
		log.Printf("Registro de cliente: %s. Registrando...", clientID)
		s.registeredClients = append(s.registeredClients, clientID)
	}
	// Unlock mutex después de modificar la lista de clientes
	s.clientMutex.Unlock()
	// Creamos y devolvemos la respuesta.
	log.Printf("Hola, %s!, hemos recibido:\n tienda=%s categoria=%s\n producto=%s precio=%v\n stock=%v fecha=%s",
		in.GetProductoId(), in.GetTienda(), in.GetCategoria(), in.GetProducto(), in.GetPrecioDescuento(), in.GetStock(), in.GetFecha())
	//msg := fmt.Sprintf("Oferta recibida para el producto ID: %s", in.GetProductoId())
	//return &pb.OfertasResponse{BrokerMessage: msg}, nil
	return nil, nil
}

func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	// 2. Creamos una nueva instancia del servidor gRPC
	s := grpc.NewServer()
	pb.RegisterOfertasServer(s, &server{
		registeredClients: make([]string, 0),
	})
	log.Printf("Servidor escuchando en %v", lis.Addr())

	// 4. Iniciamos el servidor para que empiece a aceptar peticiones en el puerto.
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
