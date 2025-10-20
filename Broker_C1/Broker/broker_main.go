package main

import (
	"context"
	"log"
	"net"
	"sync"

	// Importamos el código generado por protoc
	pb "broker_c1/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	address_broker = ":50051"
	adress_db1     = "localhost:50052"
	adress_db2     = "localhost:50053"
	adress_db3     = "localhost:50054"
)

// Definimos una struct para nuestro servidor. Debe embeber el UnimplementedGreeterServer.
// Esto asegura la compatibilidad hacia adelante si se añaden más RPCs al servicio.
type server struct {
	pb.UnimplementedOfertasServer
	registeredClients []string // Lista para almacenar los IDs de clientes registrados
	clientMutex       sync.Mutex
	dbClients         []pb.DynamoDBClient // Clientes para conectarse a los servidores de BD
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

	for i, dbClient := range s.dbClients {
		log.Printf("Enviando oferta a DB #%d...", i+1)

		storeResp, err := dbClient.GuardarOfertas(ctx, in)
		if err != nil {
			// If ANY database fails, stop and return an error to the client
			log.Printf("ERROR from DB server #%d: %v. Halting operation.", i+1, err)
			return nil, status.Errorf(codes.Internal, "Failed to store offer in DB #%d: %v", i+1, err)
		}

		log.Printf("Guardado exitosamente en DB #%d (Respuesta: %s)", i+1, storeResp.GetBrokerMessage())
	}

	log.Println("Oferta guardada en todas las bases de datos.")

	return &pb.OfertasResponse{BrokerMessage: "Oferta registrada"}, nil
}

func newDBClient(address string) (pb.DynamoDBClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return pb.NewDynamoDBClient(conn), conn, nil
}

func main() {

	// --- 6. Connect to ALL database servers ---
	dbAddresses := []string{adress_db1, adress_db2, adress_db3}
	var dbClients []pb.DynamoDBClient
	var connections []*grpc.ClientConn

	for _, addr := range dbAddresses {
		client, conn, err := newDBClient(addr)
		if err != nil {
			log.Fatalf("Did not connect to database server at %s: %v", addr, err)
		}
		defer conn.Close() // Defer closing all connections
		dbClients = append(dbClients, client)
		connections = append(connections, conn) // Keep track of connections to close them
	}
	log.Printf("Successfully connected to %d database servers.", len(dbClients))

	lis, err := net.Listen("tcp", address_broker)
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	// 2. Creamos una nueva instancia del servidor gRPC
	s := grpc.NewServer()
	pb.RegisterOfertasServer(s, &server{
		registeredClients: make([]string, 0), //Aquí se registran los clientes
		dbClients:         dbClients,
	})
	log.Printf("Servidor broker escuchando en %v", lis.Addr())

	// 4. Iniciamos el servidor para que empiece a aceptar peticiones en el puerto.
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
