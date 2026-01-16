package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "product/api/server/productinternal"
)

func main() {
	conn, err := grpc.Dial("localhost:8085", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewProductInternalServiceClient(conn)
	req := &pb.CreateProductRequest{
		Name:  "iPhone 15 Pro",
		Price: 1200.00,
		Stock: 100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.CreateProduct(ctx, req)
	if err != nil {
		log.Fatalf("could not create product: %v", err)
	}

	log.Printf("✅ Product Created Successfully!")
	log.Printf("UUID: %s", resp.ProductID)
	log.Printf("Copy this UUID to use in create_order.go")
}
