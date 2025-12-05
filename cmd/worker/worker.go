package main

import (
	"bytes"
	"distGrep/internals/mapreduce"
	"distGrep/internals/types"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

// runGrepMap executes the "map" for a single task: scan the file and emit KVs
// where Key = "file:line", Value = full line text.
// func RunGrepMap(pattern, path string) ([]types.KeyValue, error) {
// 	log.Printf("[Worker] Running map on file=%s pattern=%q", path, pattern)

// 	regPattern, err := regexp.Compile(pattern)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
// 	}

// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("error opening file %q: %w", path, err)
// 	}
// 	defer file.Close()

// 	var kvs []types.KeyValue

// 	scanner := bufio.NewScanner(file)
// 	lineNum := 0

// 	for scanner.Scan() {
// 		lineNum++
// 		text := scanner.Text()

// 		if regPattern.MatchString(text) {
// 			key := fmt.Sprintf("%s:%d", path, lineNum)
// 			kvs = append(kvs, types.KeyValue{
// 				Key:   key,
// 				Value: text,
// 			})
// 		}
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, fmt.Errorf("error reading file %q: %w", path, err)
// 	}

// 	log.Printf("[Worker] Completed map on file=%s. %d matches.", path, len(kvs))
// 	return kvs, nil
// }

// How /map router will work
func HandleMap(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return

	}

	var req types.MapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[Worker] Error decoding request: %v", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	kvs, err := mapreduce.MapPhase(req.Pattern, req.Path)
	if err != nil {
		log.Printf("[Worker] Map error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &types.MapResponse{KVs: kvs}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[Worker] Error encoding response: %v", err)
		return
	}

	log.Printf("[Worker] /map finished in %s for file=%s", time.Since(start), req.Path)
}

// How workers will register to the Master
func RegisterWorker(masterURL string, workerAddr string) {
	req := map[string]string{"address": workerAddr}
	data, _ := json.Marshal(req)

	for {
		_, err := http.Post(masterURL+"/register", "application/json", bytes.NewBuffer(data))
		if err == nil {
			log.Printf("[Worker] Registered to master at %s", masterURL)
			return
		}

		log.Printf("[Worker] Failed to register. Retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
}

func main() {
	addr := flag.String("addr", ":8001", "worker listen address")
	masterURL := flag.String("master", "http://localhost:9000", "master address")
	flag.Parse()

	log.SetPrefix("[Worker] ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// register with master
	go RegisterWorker(*masterURL, "http://localhost"+*addr)

	http.HandleFunc("/map", HandleMap)

	log.Printf("Worker running on %s ...", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
