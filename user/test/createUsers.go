package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "user/api/server/userpublicapi"
)

func main() {
	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Не удалось подключиться: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserPublicAPIClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Пользователь 1: только логин
	//resp1, err := client.CreateUser(ctx, &pb.CreateUserRequest{
	//	Login: "userWITHnotification-3",
	//})
	//if err != nil {
	//	log.Fatalf("Ошибка создания первого пользователя: %v", err)
	//}
	//log.Printf("Создан пользователь 1: %s", resp1.GetUserID())

	//// Пользователь 2: все поля
	email := "user-1@example.com"
	telegram := "@user-1"
	resp2, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Login:    "user-1",
		Email:    &email,
		Telegram: &telegram,
	})
	if err != nil {
		log.Fatalf("Ошибка создания второго пользователя: %v", err)
	}
	log.Printf("Создан пользователь 2: %s", resp2.GetUserID())
}
