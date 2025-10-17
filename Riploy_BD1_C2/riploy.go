package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	pb "riploy/proto"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
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
		precio_base, err := strconv.Atoi(record[4])
		if err != nil {
			log.Printf("Warning: could not parse precio_base '%s'. Skipping row.", record[4])
			continue
		}

		stock, err := strconv.Atoi(record[5])
		if err != nil {
			log.Printf("Warning: could not parse stock '%s'. Skipping row.", record[5])
			continue
		}

		resp, err := client.Ofertas(ctx, &pb.OfertasRequest{
			ProductoId: record[0],
			Tienda:     record[1],
			Categoria:  record[2],
			Producto:   record[3],
			PrecioBase: int32(precio_base),
			Stock:      int32(stock),
		})

		if err != nil {
			log.Printf("Error sending operation for ProductoId %s: %v", record[0], err)
			continue // Continue to the next product even if one fails
		}

		log.Printf("Response for %s: %s", record[0], resp.GetBrokerMessage())
	}

	log.Println("Finished sending all offers from CSV.")
}
