package cruds

import (
	pb "api-service/api/proto"
	"api-service/internal/pkg/logger"
	"context"
	"encoding/json"
	"net/http"
)

type CRUDOperations struct {
	tsc    pb.TaskServiceClient
	logger *logger.KafkaLogger
}

func NewCRUDOperations(tsc pb.TaskServiceClient, logger *logger.KafkaLogger) *CRUDOperations {
	return &CRUDOperations{
		tsc:    tsc,
		logger: logger,
	}
}

// Helper to decode JSON
func decodeJSON[T any](r *http.Request, dst *T) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// POST /create
func (crud *CRUDOperations) HandleCreate(w http.ResponseWriter, r *http.Request) {
	crud.logger.Logger().Info().Msg("request HandleCreate POST")
	if r.Method != http.MethodPost {
		crud.logger.Logger().Warn().Str("method", r.Method).Msg("invalid method for List")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var task pb.CreateTask
	if err := decodeJSON(r, &task); err != nil {
		crud.logger.Logger().Error().Err(err).Msg("failed to decode CreateTask")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	crud.logger.Logger().Info().Msg("RPC call CreateTask")
	_, err := crud.tsc.Create(r.Context(), &task)
	if err != nil {
		crud.logger.Logger().Error().Err(err).Msg("CreateTask RPC failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /list
func (crud *CRUDOperations) HandleList(w http.ResponseWriter, r *http.Request) {
	crud.logger.Logger().Info().Msg("request HandleList GET")
	if r.Method != http.MethodGet {
		crud.logger.Logger().Warn().Str("method", r.Method).Msg("invalid method for List")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	crud.logger.Logger().Info().Msg("RPC call List")
	tasksList, err := crud.tsc.List(context.Background(), &pb.TaskID{ID: "1"})
	if err != nil {
		crud.logger.Logger().Error().Err(err).Msg("List RPC failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	crud.logger.Logger().Info().Msg("send HandleList response")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasksList); err != nil {
		crud.logger.Logger().Error().Err(err).Msg("failed to encode JSON response")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// DELETE /delete
func (crud *CRUDOperations) HandleDelete(w http.ResponseWriter, r *http.Request) {
	crud.logger.Logger().Info().Msg("request HandleDelete DELETE")
	if r.Method != http.MethodDelete {
		crud.logger.Logger().Warn().Str("method", r.Method).Msg("invalid method for Delete")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskID pb.TaskID
	if err := decodeJSON(r, &taskID); err != nil {
		crud.logger.Logger().Error().Err(err).Msg("failed to decode TaskID")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	crud.logger.Logger().Info().Msg("RPC call DeleteTask")
	_, err := crud.tsc.Delete(r.Context(), &taskID)
	if err != nil {
		crud.logger.Logger().Error().Err(err).Msg("DeleteTask RPC failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PUT /done
func (crud *CRUDOperations) HandleDone(w http.ResponseWriter, r *http.Request) {
	crud.logger.Logger().Info().Msg("request HandleDone PUT")
	if r.Method != http.MethodPut {
		crud.logger.Logger().Warn().Str("method", r.Method).Msg("invalid method for Done")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var taskID pb.TaskID
	if err := decodeJSON(r, &taskID); err != nil {
		crud.logger.Logger().Error().Err(err).Msg("failed to decode TaskID")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	crud.logger.Logger().Info().Msg("RPC call Done")
	_, err := crud.tsc.Done(r.Context(), &taskID)
	if err != nil {
		crud.logger.Logger().Error().Err(err).Msg("Done RPC failed")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
