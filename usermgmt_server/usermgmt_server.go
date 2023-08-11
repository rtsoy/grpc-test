package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"

	pb "github.com/rtsoy/grpc-test/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	port     = ":50051"
	filename = "users.json"
)

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
}

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{}
}

func (s *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())

	createdUser := &pb.User{
		Id:   rand.Int31(),
		Name: in.GetName(),
		Age:  in.GetAge(),
	}

	usersList := &pb.UsersList{}

	readBytes, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("file `%s` not found. creating a new file...\n", filename)

			usersList.Users = append(usersList.Users, createdUser)

			jsonBytes, err := protojson.Marshal(usersList)
			if err != nil {
				log.Fatalf("JSON Marshalling failed: %v", err)
			}

			if err := os.WriteFile(filename, jsonBytes, 0664); err != nil {
				log.Fatalf("failed write to file: %v", err)
			}

			return createdUser, nil
		}

		log.Fatalf("failed to read a file: %v", err)
	}

	if err := protojson.Unmarshal(readBytes, usersList); err != nil {
		log.Fatal("JSON Unmarshalling failed")
	}

	usersList.Users = append(usersList.Users, createdUser)

	jsonBytes, err := protojson.Marshal(usersList)
	if err != nil {
		log.Fatalf("JSON Marshalling failed: %v", err)
	}

	if err := os.WriteFile(filename, jsonBytes, 0664); err != nil {
		log.Fatalf("failed write to file: %v", err)
	}

	return createdUser, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read from file: %v", err)
	}

	usersList := &pb.UsersList{}
	if err := protojson.Unmarshal(jsonBytes, usersList); err != nil {
		log.Fatalf("fJSON Unmarshalling failed")
	}

	return usersList, nil
}

func (s *UserManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterUserManagementServer(server, s)
	log.Printf("server is listening at %v", lis.Addr())

	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func main() {
	server := NewUserManagementServer()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
