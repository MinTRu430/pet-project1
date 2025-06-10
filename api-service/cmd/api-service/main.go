package main

import (
	pb "api-service/api/proto"
	"api-service/internal/cruds"
	"api-service/internal/pkg/logger"
	"log"
	"net/http"

	"google.golang.org/grpc"
)

func main() {
	grpcConn, err := grpc.Dial("localhost:8081", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer grpcConn.Close()

	logger := logger.NewKafkaLogger("api-service")
	defer logger.Close()
	logger.Logger().Info().Msg("start logger! api-service")

	taskManager := pb.NewTaskServiceClient(grpcConn)

	u := cruds.NewCRUDOperations(taskManager, logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/create", u.HandleCreate)
	mux.HandleFunc("/list", u.HandleList)
	mux.HandleFunc("/delete", u.HandleDelete)
	mux.HandleFunc("/done", u.HandleDone)

	listedAddr := ":8080"
	log.Printf("starting listining server at %s", listedAddr)
	http.ListenAndServe(listedAddr, mux)

}
