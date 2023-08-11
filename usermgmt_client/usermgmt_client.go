package main

import (
	"context"
	"log"
	"time"

	pb "github.com/rtsoy/grpc-test/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const address = "localhost:50051"

func main() {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewUserManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	newUsers := map[string]int32{
		"Alice": 32,
		"Mark":  17,
		"John":  25,
		"Marry": 20,
	}

	for name, age := range newUsers {
		resp, err := c.CreateNewUser(ctx, &pb.NewUser{
			Name: name,
			Age:  age,
		})
		if err != nil {
			log.Fatalf("failed to create an user: %v", err)
		}

		log.Printf("USER :: ID=%d, NAME=%s, AGE=%d\n", resp.GetId(), resp.GetName(), resp.GetAge())
	}

	response, err := c.GetUsers(ctx, &pb.GetUsersParams{})
	if err != nil {
		log.Fatalf("failed to retrieve users: %v", err)
	}

	log.Printf("USERS LIST :: %v\n", response.GetUsers())
}
