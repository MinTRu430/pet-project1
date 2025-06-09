package main

import (
	pb "api-service/api/proto"
	"api-service/internal/cruds"
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

	taskManager := pb.NewTaskServiceClient(grpcConn)

	u := cruds.NewCRUDOperations(taskManager)

	mux := http.NewServeMux()

	mux.HandleFunc("/create", u.Create)
	mux.HandleFunc("/list", u.List)
	mux.HandleFunc("/delete", u.Delete)
	mux.HandleFunc("/done", u.Done)

	listedAddr := ":8080"
	log.Printf("starting listining server at %s", listedAddr)
	http.ListenAndServe(listedAddr, mux)

}
