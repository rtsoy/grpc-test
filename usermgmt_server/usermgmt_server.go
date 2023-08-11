package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	pb "github.com/rtsoy/grpc-test/usermgmt"
	"google.golang.org/grpc"
)

const port = ":50051"

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
	userList *pb.UsersList
}

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{
		userList: &pb.UsersList{},
	}
}

func (s *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())

	createdUser := &pb.User{
		Id:   rand.Int31(),
		Name: in.GetName(),
		Age:  in.GetAge(),
	}

	s.userList.Users = append(s.userList.Users, createdUser)

	return createdUser, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {
	return s.userList, nil
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
