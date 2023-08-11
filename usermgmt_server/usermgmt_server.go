package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	pb "github.com/rtsoy/grpc-test/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	grpcPort = os.Getenv("GRPC_PORT")

	host     = os.Getenv("POSTGRES_HOST")
	port     = os.Getenv("POSTGRES_PORT")
	user     = os.Getenv("POSTGRES_USER")
	password = os.Getenv("POSTGRES_PASSWORD")
	dbName   = os.Getenv("POSTGRES_DB")

	ctx = context.Background()
)

type UserManagementServer struct {
	conn *pgx.Conn
	pb.UnimplementedUserManagementServer
}

func NewUserManagementServer(conn *pgx.Conn) *UserManagementServer {
	return &UserManagementServer{
		conn: conn,
	}
}

func (s *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start a transaction: %v", err)
	}

	var userID int32
	if err := tx.QueryRow(ctx, `
		INSERT INTO users(name, age)
		VALUES ($1, $2)
		RETURNING id
    `, in.GetName(), in.GetAge()).Scan(&userID); err != nil {
		return nil, fmt.Errorf("failed to execute a query: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit a transaction: %v", err)
	}

	createdUser := &pb.User{
		Id:   userID,
		Name: in.GetName(),
		Age:  in.GetAge(),
	}

	return createdUser, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UsersList, error) {
	rows, err := s.conn.Query(ctx, `
		SELECT id, name, age
		FROM users
    `)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to retrieve a data: %v", err)
	}
	defer rows.Close()

	usersList := &pb.UsersList{}

	for rows.Next() {
		user := &pb.User{}
		if err := rows.Scan(&user.Id, &user.Name, &user.Age); err != nil {
			return nil, fmt.Errorf("failed to read a data from db: %v", err)
		}

		usersList.Users = append(usersList.Users, user)
	}

	return usersList, nil
}

func (s *UserManagementServer) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	reflection.Register(server)
	pb.RegisterUserManagementServer(server, s)

	log.Printf("server is listening at %v", lis.Addr())

	if err := server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func main() {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)

	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		log.Fatalf("failed to establish a connection to postgres: %v", err)
	}
	defer conn.Close(ctx)

	server := NewUserManagementServer(conn)

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
