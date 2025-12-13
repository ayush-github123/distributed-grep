package master

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	
	"distGrep/internals/grep"
	"distGrep/internals/mapreduce"
	"distGrep/internals/types"
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
	if r.Method != http.MethodPost {
		http.Error(w, "Mehod not allowed", http.StatusMethodNotAllowed)
	}

	var req types.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error while decoding", http.StatusInternalServerError)
	}

	pattern := req.Pattern
	if pattern == "" {
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

	// tasks := make([]mapreduce.Task, len(files))
	var tasks []mapreduce.Task	//slice can grow sizes with increasing data in it
	for _, f := range files {
		fileInfo, err := os.Stat(f)
		if err!=nil{
			fmt.Printf("Error reading the file, %v", err)
		}

		if fileInfo.IsDir(){
			HandleDirs(f, &tasks)
			// fmt.Printf("Its a directory")
		}else{
			// tasks[i] = mapreduce.Task{Path: f}	//now useless as we are reading dir inside dir and all share the same slice:  tasks
			tasks = append(tasks, mapreduce.Task{Path: f})

		}
	}
	// fmt.Printf("TASKS ARE : %v\n", tasks) 
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


func HandleDirs(path string, tasks *[]mapreduce.Task) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Print(err)
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())	//read full path instead of just file name
		
		if entry.IsDir() {		//if there is a dir inside a dir then again run the func -> recursive call :)
			HandleDirs(entry.Name(), tasks)
		} else {
			// fmt.Printf("%v\n", entry.Name())
			*tasks = append(*tasks, mapreduce.Task{Path: fullPath})
		}
	}
	return nil

}