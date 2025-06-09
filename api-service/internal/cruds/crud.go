package cruds

import (
	pb "api-service/api/proto"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type CRUDOperations struct {
	tsc pb.TaskServiceClient
}

func NewCRUDOperations(tsc pb.TaskServiceClient) *CRUDOperations {
	return &CRUDOperations{
		tsc: tsc,
	}
}

// POST /create
func (crud *CRUDOperations) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Println("Method is not method POST: ", r.Method)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	task := &pb.CreateTask{}
	err := json.NewDecoder(r.Body).Decode(task)
	if err != nil {
		log.Println("can't decode json:", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	_, err = crud.tsc.Create(
		context.Background(),
		task,
	)
	if err != nil {
		log.Printf("CreateTask failed: %v", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	crud.List(w, r)
}

// GET /list
func (crud *CRUDOperations) List(w http.ResponseWriter, r *http.Request) {
	tasksList, err := crud.tsc.List(context.Background(), &pb.TaskID{ID: "1"})
	if err != nil {
		log.Println("crud.tsc.List encode error:", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&tasksList); err != nil {
		log.Println("JSON encode error:", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

}

// DELETE /delete
func (crud *CRUDOperations) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		log.Println("Method is not method DELETE: ", r.Method)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	taskId := &pb.TaskID{}
	err := json.NewDecoder(r.Body).Decode(taskId)
	if err != nil {
		log.Println("can't decode json:", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	_, err = crud.tsc.Delete(
		context.Background(),
		taskId)
	if err != nil {
		log.Printf("DeleteTask failed: %v", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	crud.List(w, r)
}

// PUT /done
func (crud *CRUDOperations) Done(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		log.Println("Method is not method PUT: ", r.Method)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	taskId := &pb.TaskID{}
	err := json.NewDecoder(r.Body).Decode(taskId)
	if err != nil {
		log.Println("can't decode json:", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	_, err = crud.tsc.Done(
		context.Background(),
		taskId)
	if err != nil {
		log.Printf("DeleteTask failed: %v", err)
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	crud.List(w, r)
}
