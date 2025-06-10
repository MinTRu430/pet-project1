package taskmanager

import (
	"context"
	"database/sql"
	pb "db-service/api/proto"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"
)

type TaskManager struct {
	pb.UnimplementedTaskServiceServer
	db          *sql.DB
	redisClient *redis.Client
	kafkaLogger zerolog.Logger
}

func NewTaskManager(db *sql.DB, redisClient *redis.Client, logger zerolog.Logger) *TaskManager {
	return &TaskManager{
		db:          db,
		redisClient: redisClient,
		kafkaLogger: logger,
	}
}

func (tm *TaskManager) Create(ctx context.Context, in *pb.CreateTask) (*pb.Nothing, error) {
	tm.kafkaLogger.Info().Str("header", in.Header).Str("body", in.Body).Msg("received Create request")

	if in.Header == "" && in.Body == "" {
		return nil, fmt.Errorf("header is nil: %s or body is nil: %s", in.Header, in.Body)
	}

	tm.kafkaLogger.Info().Msg("query insert into tasks")
	_, err := tm.db.Exec("INSERT INTO tasks(header, body) VALUES ($1, $2)", in.Header, in.Body)
	if err != nil {
		tm.kafkaLogger.Error().Err(err).Msg("insert into tasks insert error")
		return &pb.Nothing{Dummy: false}, fmt.Errorf("insert into tasks insert error %s", err)
	}

	if err := tm.redisClient.Del(ctx, "task_list").Err(); err != nil {
		tm.kafkaLogger.Warn().Err(err).Msg("failed to delete task_list from Redis")
	} else {
		tm.kafkaLogger.Info().Msg("deleted task_list from Redis")
	}

	tm.kafkaLogger.Info().Msg("task successfully inserted into DB")
	return &pb.Nothing{Dummy: false}, nil
}

func (tm *TaskManager) List(ctx context.Context, in *pb.TaskID) (*pb.TaskList, error) {
	tm.kafkaLogger.Info().Msg("received List request")

	cacheKey := "task_list"

	tm.kafkaLogger.Info().Msg("attempting to get task_list from Redis cache")
	val, err := tm.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var cachedTasks pb.TaskList
		if jsonErr := json.Unmarshal([]byte(val), &cachedTasks); jsonErr == nil {
			tm.kafkaLogger.Info().Msg("get TaskList from Redis cache")
			return &cachedTasks, nil
		}
	}

	tm.kafkaLogger.Info().Msg("query select from tasks")
	rows, err := tm.db.Query("SELECT id, header, body, isdone FROM tasks")
	if err != nil {
		tm.kafkaLogger.Error().Err(err).Msg("select from tasks error")

		return nil, err
	}
	defer rows.Close()

	var tasks []*pb.Task
	for rows.Next() {
		var t pb.Task
		err := rows.Scan(&t.ID, &t.Header, &t.Body, &t.IsDone)
		if err != nil {
			tm.kafkaLogger.Error().Err(err).Msg("rows scan error")
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	result := &pb.TaskList{Tasks: tasks}

	data, err := json.Marshal(result)
	if err != nil {
		tm.kafkaLogger.Error().Err(err).Msg("can't marshal List")
		return nil, fmt.Errorf("can't marshal List, %v", err)
	}

	tm.redisClient.Set(ctx, cacheKey, data, 1*time.Minute)
	tm.kafkaLogger.Info().Msg("cached task_list to Redis for 1 minute")

	return result, nil
}
func (tm *TaskManager) Delete(ctx context.Context, in *pb.TaskID) (*pb.Nothing, error) {
	tm.kafkaLogger.Info().Str("id", in.ID).Msg("received Delete request")
	if in.ID == "" {
		tm.kafkaLogger.Warn().Msg("empty ID provided in Delete")
		return nil, fmt.Errorf("id is empty %s", in.ID)
	}

	tm.kafkaLogger.Info().Str("id", in.ID).Msg("executing delete from DB")
	_, err := tm.db.Exec("DELETE FROM tasks WHERE id = $1;", in.ID)
	if err != nil {
		tm.kafkaLogger.Error().Err(err).Str("id", in.ID).Msg("failed to delete task")
		return &pb.Nothing{Dummy: false}, fmt.Errorf("delete error %s", err)
	}

	if err := tm.redisClient.Del(ctx, "task_list").Err(); err != nil {
		tm.kafkaLogger.Warn().Err(err).Msg("failed to delete task_list from Redis")
	} else {
		tm.kafkaLogger.Info().Msg("deleted task_list from Redis")
	}

	return &pb.Nothing{Dummy: false}, nil
}
func (tm *TaskManager) Done(ctx context.Context, in *pb.TaskID) (*pb.Nothing, error) {
	tm.kafkaLogger.Info().Str("id", in.ID).Msg("received Done request")

	if in.ID == "" {
		tm.kafkaLogger.Warn().Msg("empty ID provided in Done")
		return nil, fmt.Errorf("id is empty %s", in.ID)
	}

	tm.kafkaLogger.Info().Str("id", in.ID).Msg("marking task as done")
	_, err := tm.db.Exec("UPDATE tasks SET isdone = true WHERE id = $1;", in.ID)
	if err != nil {
		tm.kafkaLogger.Error().Err(err).Str("id", in.ID).Msg("failed to mark task as done")
		return &pb.Nothing{Dummy: false}, fmt.Errorf("update error %s", err)
	}

	tm.kafkaLogger.Info().Str("id", in.ID).Msg("task successfully marked as done")
	if err := tm.redisClient.Del(ctx, "task_list").Err(); err != nil {
		tm.kafkaLogger.Warn().Err(err).Msg("failed to delete task_list from Redis")
	} else {
		tm.kafkaLogger.Info().Msg("deleted task_list from Redis")
	}

	return &pb.Nothing{Dummy: false}, nil
}
