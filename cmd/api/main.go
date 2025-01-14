package main

import (
	"context"
	"log"
	"net"
	"time"

	"app/api/proto"
	"app/internal/api/file"
	fileDB "app/internal/api/file/db"
	"app/internal/config"
	postgresqlClient "app/pkg/client/postgresql"
	"app/pkg/logging"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.GetConfig()
	logger := logging.NewLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	postgreSQLClient, err := postgresqlClient.NewClient(logger, ctx, 4, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.Username, cfg.Postgres.Password, cfg.Postgres.Database)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	fileRepository := fileDB.NewRepository(logger, postgreSQLClient)
	startGRPCServer(logger, cfg, fileRepository)
}

func startGRPCServer(logger *logging.Logger, cfg *config.Config, fileRepository file.FileRepository) {
	lis, err := net.Listen("tcp", cfg.Listen.GRPC.Host+":"+cfg.Listen.GRPC.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	srv := file.NewServer(logger, fileRepository)
	proto.RegisterFileServiceServer(grpcServer, srv)

	log.Println("gRPC server is running on port :" + cfg.Listen.GRPC.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
