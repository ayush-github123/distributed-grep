package master

import (
	"distGrep/internals/grep"
	"distGrep/internals/mapreduce"
	"distGrep/internals/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	workers  []string
	workerMu sync.Mutex
)

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	type registerReq struct {
		Address string `json:"address"`
	}
	var req registerReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}

	workerMu.Lock()
	workers = append(workers, req.Address)
	workerMu.Unlock()

	log.Printf("[MASTER] Registered worker: %s", req.Address)
	w.WriteHeader(http.StatusOK)
}

func GetWorkers() []string {
	workerMu.Lock()
	defer workerMu.Unlock()
	return append([]string{}, workers...)
}


func HandleRunJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		http.Error(w, "Mehod not allowed", http.StatusMethodNotAllowed)
	}

	var req types.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err!=nil{
		http.Error(w, "Error while decoding", http.StatusInternalServerError)
	}

	pattern := req.Pattern
	if pattern == ""{
		http.Error(w, "Enter the pattern to GREP", http.StatusServiceUnavailable)
		return
	}

	files := req.Files
	if len(files) == 0 {
		http.Error(w, "no files available", http.StatusServiceUnavailable)
		return
	}

	workers := GetWorkers()
	if len(workers) == 0 {
		http.Error(w, "no workers available", http.StatusServiceUnavailable)
		return
	}
	
	tasks := make([]mapreduce.Task, len(files))
	for i, f := range files{
		tasks[i] = mapreduce.Task{Path: f}
	}
	// fmt.Println(tasks)
	job := mapreduce.NewJob(pattern, &grep.GrepReducer{}, tasks, workers)

	results, err := job.Execute()
	if err != nil {
		fmt.Println("\nexecution error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := types.RunResponse{Results: results}
	json.NewEncoder(w).Encode(resp)
}