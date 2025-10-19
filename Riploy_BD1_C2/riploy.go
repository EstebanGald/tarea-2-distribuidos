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

	pb "riploy/proto"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {

	//random_seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//Conectar al servidor gRPC
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error al conectar al servidor: %v", err)
	}
	defer conn.Close()

	client := pb.NewOfertasClient(conn)

	ctx := context.Background()

	// 2. Open the CSV file
	file, err := os.Open("riploy_catalogo.csv")
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// 3. Create a new CSV reader
	reader := csv.NewReader(file)
	// Skip the header row
	if _, err := reader.Read(); err != nil {
		log.Fatalf("Error reading header from CSV: %v", err)
	}

	// 4. Loop through each record in the CSV
	for {
		record, err := reader.Read()
		// Stop at the end of the file
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading record from CSV: %v", err)
		}

		// 5. Parse and convert data for the request
		// The fields precio_base and stock are defined as int32 in the proto file.
		// The CSV provides these as strings, so they must be converted.
		// Parse original price
		originalPrecioBase, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Warning: could not parse precio_base '%s'. Skipping row.", record[4])
			continue
		}

		stock, err := strconv.Atoi(record[5])
		if err != nil {
			log.Printf("Warning: could not parse stock '%s'. Skipping row.", record[5])
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
			ClienteId:       "Riploy", // Identificador del cliente
		})

		if err != nil {
			log.Printf("Error sending operation for ProductoId %s: %v", record[0], err)
			continue // Continuar a la siguiente fila incluso si hay un error
		}

		log.Printf("Response for %s: %s", record[0], resp.GetBrokerMessage())
	}

	log.Println("Finished sending all offers from CSV.")
}
