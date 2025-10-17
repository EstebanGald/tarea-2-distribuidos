package main

import (
	"context"
	"fmt"
	"log"
	"net"

	// Importamos el código generado por protoc
	pb "broker_main/proto" // Reemplaza con el path de tu módulo

	"google.golang.org/grpc"
)

// Definimos una struct para nuestro servidor. Debe embeber el UnimplementedGreeterServer.
// Esto asegura la compatibilidad hacia adelante si se añaden más RPCs al servicio.
type server struct {
	pb.UnimplementedOfertasServer
}

// SayHello es la implementación de la función definida en el archivo .proto.
// Esta es la lógica real que se ejecuta cuando un cliente llama a este RPC.
func (s *server) Ofertas(ctx context.Context, in *pb.OfertasRequest) (*pb.OfertasResponse, error) {
	log.Printf("Recibida petición de: %v", in.GetProductoId())
	// Creamos y devolvemos la respuesta.
	log.Printf("Hola, %s!, hemos recibido: tienda=%s categoria=%s producto=%s precio=%v stock=%v",
		in.GetProductoId(), in.GetTienda(), in.GetCategoria(), in.GetProducto(), in.GetPrecioBase(), in.GetStock())
	msg := fmt.Sprintf("Oferta recibida para el producto ID: %s", in.GetProductoId())
	return &pb.OfertasResponse{BrokerMessage: msg}, nil
}

func main() {

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Fallo al escuchar: %v", err)
	}

	// 2. Creamos una nueva instancia del servidor gRPC
	s := grpc.NewServer()
	pb.RegisterOfertasServer(s, &server{})
	log.Printf("Servidor escuchando en %v", lis.Addr())

	// 4. Iniciamos el servidor para que empiece a aceptar peticiones en el puerto.
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Fallo al servir: %v", err)
	}
}
