package main

import (
	"context"
	"log"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Replace 'grpc-profile' with your module name from go.mod
	pb "grpc-profile/pb"
)

type ProfileDoc struct {
	UserID   string   `bson:"user_id"`
	FullName string   `bson:"full_name"`
	Email    string   `bson:"email"`
	Skills   []string `bson:"skills"`
}

// This is the gRPC server struct
type server struct {
	pb.UnimplementedProfileServiceServer
	col *mongo.Collection
}

// GetUserProfile implements the gRPC service
func (s *server) GetUserProfile(ctx context.Context, in *pb.UserRequest) (*pb.UserProfileResponse, error) {
	var result ProfileDoc

	// Filter by the userID sent in the gRPC request
	filter := bson.M{"user_id": in.GetUserId()}

	err := s.col.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, "User not found")
		}
		return nil, status.Errorf(codes.Internal, "Database error: %v", err)
	}

	// Map MongoDB result to gRPC Protobuf response
	return &pb.UserProfileResponse{
		UserId:   result.UserID,
		FullName: result.FullName,
		Email:    result.Email,
		Skills:   result.Skills,
	}, nil
}

func connectDB() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	return client.Database("user_db").Collection("profiles")
}

func main() {
	// 1. Connect to MongoDB
	collection := connectDB()
	log.Println("Connected to MongoDB...")

	// 2. Create a TCP listener on port 50051
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 3. Create and Register the gRPC server
	s := grpc.NewServer()
	pb.RegisterProfileServiceServer(s, &server{col: collection})

	log.Println("gRPC Server is running on port 50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

