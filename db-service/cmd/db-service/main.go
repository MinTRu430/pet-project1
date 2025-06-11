package main

import (
	"database/sql"
	messagepb "db-service/api/proto"
	"db-service/internal/pkg/logger"
	"db-service/internal/taskmanager"
	"log"
	"net"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	dsn := "host=127.0.0.1 port=5432 user=dude password=pass dbname=tasksdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("1failed to connect to database: %v", err)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "redkaPass",
	})

	err = db.Ping()
	if err != nil {
		log.Fatalf("2failed to connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalln("cant listen port", err)
	}

	logger := logger.NewKafkaLogger("db-service")
	defer logger.Close()
	logger.Logger().Info().Msg("start-logging db-service!!!")

	server := grpc.NewServer()

	messagepb.RegisterTaskServiceServer(server, taskmanager.NewTaskManager(db, rdb, logger))

	log.Println("gRPC server listening on :8081")
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
