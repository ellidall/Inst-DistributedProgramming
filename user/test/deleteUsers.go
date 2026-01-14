package main

//import (
//	"context"
//	"time"
//
//	log "github.com/sirupsen/logrus"
//	"google.golang.org/grpc"
//
//	pb "user/api/server/userpublicapi"
//)
//
//func main() {
//	conn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
//	if err != nil {
//		log.Fatalf("Не удалось подключиться: %v", err)
//	}
//	defer conn.Close()
//
//	client := pb.NewUserPublicAPIClient(conn)
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	// Удаляем первого пользователя (soft delete)
//	_, err = client.DeleteUser(ctx, &pb.DeleteUserRequest{
//		UserID: "019bb3d1-ce50-7dac-a057-2ca737e72ef4", // Замени на реальный ID
//	})
//	if err != nil {
//		log.Fatalf("Ошибка удаления первого пользователя: %v", err)
//	}
//	log.Println("Первый пользователь удалён (soft delete)")

//// Удаляем второго пользователя (soft delete)
//_, err = client.DeleteUser(ctx, &pb.DeleteUserRequest{
//	UserID: "019bb366-c0cc-72ab-aaba-26a4a35797a4", // Замени на реальный ID
//})
//if err != nil {
//	log.Fatalf("Ошибка удаления второго пользователя: %v", err)
//}
//log.Println("Второй пользователь удалён (soft delete)")
//}
