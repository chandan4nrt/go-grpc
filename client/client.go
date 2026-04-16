package main

import (
	"context"
	"log"
	"time"

	pb "grpc-profile/pb"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewProfileServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetUserProfile(ctx, &pb.UserRequest{
		UserId: "123",
	})
	if err != nil {
		log.Fatalf("could not get profile: %v", err)
	}

	log.Printf("User: %s, Email: %s, Skills: %v",
		resp.FullName, resp.Email, resp.Skills)
}
