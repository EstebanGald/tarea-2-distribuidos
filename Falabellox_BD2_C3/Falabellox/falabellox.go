package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	pb "falabellox_bd2_c3/proto"

	"google.golang.org/grpc"
)

const (
	address_broker = "localhost:50051"
)

func main() {

	//random seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//Conectar al servidor gRPC
	conn, err := grpc.Dial(address_broker, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewOfertasClient(conn)

	ctx := context.Background()

	// 2. Open the CSV file
	file, err := os.Open("falabellox_catalogo.csv")
	if err != nil {
		log.Fatalf("Error al abrir CSV: %v", err)
	}
	defer file.Close()

	// 3. Create a new CSV reader
	reader := csv.NewReader(file)
	// Skip the header row
	if _, err := reader.Read(); err != nil {
		log.Fatalf("Error leyendo header CSV: %v", err)
	}

	// 4. Loop through each record in the CSV
	for {
		record, err := reader.Read()
		// Stop at the end of the file
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error leyendo oferta de CSV: %v", err)
		}

		//convertir precio y stock a int32
		originalPrecioBase, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Warning: no se pudo transformar precio_base '%s'. Saltando fila.", record[4])
			continue
		}

		stock, err := strconv.Atoi(record[5])
		if err != nil {
			log.Printf("Warning: no se pudo transformar stock '%s'. Saltando fila.", record[5])
			continue
		}

		//Aplicar descuento aleatorio entre 10% y 50%
		discountPercent := 0.10 + r.Float64()*0.40
		originalPrecioFloat := float64(originalPrecioBase)
		discountedPrecioFloat := originalPrecioFloat * (1.0 - discountPercent)
		finalPrecio := int32(discountedPrecioFloat)

		//Get Fecha Actual
		currentTime := time.Now()
		formattedDate := currentTime.Format("2006-01-02")

		resp, err := client.Ofertas(ctx, &pb.OfertasRequest{
			ProductoId:      record[0],
			Tienda:          record[1],
			Categoria:       record[2],
			Producto:        record[3],
			PrecioDescuento: int32(finalPrecio),
			Stock:           int32(stock),
			Fecha:           formattedDate,
			ClienteId:       "Falabellox", // Identificador del cliente
		})

		if err != nil {
			log.Printf("Error enviando operacion para ProductoId %s: %v", record[0], err)
			continue // Continuar a la siguiente fila incluso si hay un error
		}

		log.Printf("Respuesta de %s: %s", record[0], resp.GetBrokerMessage())

		// Esperar un tiempo aleatorio entre 500ms y 2000ms antes de enviar la siguiente oferta
		sleepDuration := time.Duration(500+r.Intn(1500)) * time.Millisecond
		time.Sleep(sleepDuration)
	}

	log.Println("Se han enviado todas las ofertas del CSV.")
}
