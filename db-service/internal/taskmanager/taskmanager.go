package taskmanager

import (
	"context"
	"database/sql"
	pb "db-service/api/proto"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)

type TaskManager struct {
	pb.UnimplementedTaskServiceServer
	db          *sql.DB
	redisClient *redis.Client
}

func NewTaskManager(db *sql.DB, redisClient *redis.Client) *TaskManager {
	return &TaskManager{
		db:          db,
		redisClient: redisClient,
	}
}

func (tm *TaskManager) Create(ctx context.Context, in *pb.CreateTask) (*pb.Nothing, error) {
	if in.Header == "" && in.Body == "" {
		return nil, fmt.Errorf("header is nil: %s or body is nil: %s", in.Header, in.Body)
	}

	_, err := tm.db.Exec("INSERT INTO tasks(header, body) VALUES ($1, $2)", in.Header, in.Body)
	if err != nil {
		log.Printf("insert into tasks insert error %s\n", err)
		return &pb.Nothing{Dummy: false}, fmt.Errorf("insert into tasks insert error %s", err)
	}

	_ = tm.redisClient.Del(ctx, "task_list").Err()

	return &pb.Nothing{Dummy: false}, nil
}

func (tm *TaskManager) List(ctx context.Context, in *pb.TaskID) (*pb.TaskList, error) {
	cacheKey := "task_list"

	val, err := tm.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedTasks pb.TaskList
		if jsonErr := json.Unmarshal([]byte(val), &cachedTasks); jsonErr == nil {
			log.Println("get form Redis")
			return &cachedTasks, nil
		}
	}

	rows, err := tm.db.Query("SELECT id, header, body, isdone FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var t pb.Task
		err := rows.Scan(&t.ID, &t.Header, &t.Body, &t.IsDone)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	result := &pb.TaskList{Tasks: tasks}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("can't marshal List, %v", err)
	}
	redisStatus := tm.redisClient.Set(ctx, cacheKey, data, 1*time.Minute)
	log.Printf("redisStatusFullName %v", redisStatus.FullName())

	return result, nil
}
func (tm *TaskManager) Delete(ctx context.Context, in *pb.TaskID) (*pb.Nothing, error) {
	if in.ID == "" {
		return nil, fmt.Errorf("id is empty %s", in.ID)
	}

	_, err := tm.db.Exec("DELETE FROM tasks WHERE id = $1;", in.ID)
	if err != nil {
		log.Printf("delete %s\n", err)
		return &pb.Nothing{Dummy: false}, fmt.Errorf("delete error %s", err)
	}

	_ = tm.redisClient.Del(ctx, "task_list").Err()

	return &pb.Nothing{Dummy: false}, nil
}
func (tm *TaskManager) Done(ctx context.Context, in *pb.TaskID) (*pb.Nothing, error) {
	if in.ID == "" {
		return nil, fmt.Errorf("id is empty %s", in.ID)
	}

	_, err := tm.db.Exec("UPDATE tasks SET isdone = true WHERE id = $1;", in.ID)
	if err != nil {
		log.Printf("update %s\n", err)
		return &pb.Nothing{Dummy: false}, fmt.Errorf("update error %s", err)
	}

	_ = tm.redisClient.Del(ctx, "task_list").Err()

	return &pb.Nothing{Dummy: false}, nil
}
