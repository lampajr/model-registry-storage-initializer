package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lampajr/model-registry-storage-initializer/pkg/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	mlmdHostEnv     = "MLMD_HOSTNAME"
	mlmdPortEnv     = "MLMD_PORT"
	mlmdHostDefault = "localhost"
	mlmdPortDefault = "9090"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: ./mr-storage-initializer <src-uri> <dest-path>")
	}

	sourceUri := os.Args[1]
	destPath := os.Args[2]

	log.Printf("Initializing, args: src_uri [%s] dest_path[ [%s]\n", sourceUri, destPath)

	mlmdHost, ok := os.LookupEnv(mlmdHostEnv)
	if !ok || mlmdHost == "" {
		mlmdHost = mlmdHostDefault
	}
	mlmdPort, ok := os.LookupEnv(mlmdPortEnv)
	if !ok || mlmdPort == "" {
		mlmdPort = mlmdPortDefault
	}

	mlmdAddr := fmt.Sprintf("%s:%s", mlmdHost, mlmdPort)

	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := grpc.DialContext(
		ctxTimeout,
		mlmdAddr,
		grpc.WithReturnConnectionError(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("error dialing connection to mlmd server %s: %v", mlmdAddr, err)
	}

	provider, err := storage.NewModelRegistryProvider(conn)
	if err != nil {
		log.Fatalf("Error initiliazing model registry provider: %v", err)
	}

	if err := provider.DownloadModel(destPath, "", sourceUri); err != nil {
		log.Fatalf(err.Error())
	}
}
