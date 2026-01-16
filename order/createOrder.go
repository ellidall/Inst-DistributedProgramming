package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "order/api/server/orderinternalapi"
)

func main() {
	productID := "019bc26c-304a-7449-a2a9-7e4efc5eafe4"

	customerID := "ВСТАВЬ_СЮДА_UUID_ПРОДУКТА"

	conn, err := grpc.Dial("localhost:8084", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrderInternalAPIClient(conn)

	req := &pb.CreateOrderRequest{
		CustomerID: customerID,
		Items: []*pb.CreateOrderItem{
			{
				ProductID: productID,
				Count:     1,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.CreateOrder(ctx, req)
	if err != nil {
		log.Fatalf("could not create order: %v", err)
	}

	log.Printf("🚀 Order Created Successfully!")
	log.Printf("OrderID: %s", resp.OrderID)
	log.Printf("Check RabbitMQ and Grafana to see the Saga in action!")
}
