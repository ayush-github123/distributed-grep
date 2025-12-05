package mapreduce

import (
	"bufio"
	"distGrep/internals/types"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
)

type KeyValue struct {
	Key   string
	Value string
}

// type Mapper interface {
// 	Map(task Task) ([]KeyValue, error)
// }

type Reducer interface {
	Reduce(key string, values []string) (string, error)
}

type Task struct {
	Path string
}

type Job struct {
	// Mapper  Mapper
	Pattern string
	Reducer Reducer
	Tasks   []Task
	Workers []string
	logger  *log.Logger
}

func NewJob(pattern string, reducer Reducer, tasks []Task, workers []string) *Job {
	logFile, err := os.OpenFile("distgrep.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error creating log file: %v", err)
	}

	logger := log.New(logFile, "[DistGrep] ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Job{
		// Mapper:  mapper,
		Pattern: pattern,
		Reducer: reducer,
		Tasks:   tasks,
		Workers: workers,
		logger:  logger,
	}
}

func MapPhase(pattern, path string) ([]types.KeyValue, error) {
	log.Printf("[Worker] Running map on file=%s pattern=%q", path, pattern)

	regPattern, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", path, err)
	}
	defer file.Close()

	var kvs []types.KeyValue

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()

		if regPattern.MatchString(text) {
			key := fmt.Sprintf("%s:%d", path, lineNum)
			kvs = append(kvs, types.KeyValue{
				Key:   key,
				Value: text,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %q: %w", path, err)
	}

	log.Printf("[Worker] Completed map on file=%s. %d matches.", path, len(kvs))
	return kvs, nil
}

func (j *Job) Execute() (map[string]string, error) {
	j.logger.Println("Starting MapReduce Job")

	// Map Phase
	j.logger.Println("Starting Map Phase")
	var mu sync.Mutex
	intermediates := make(map[string][]string)

	var wg sync.WaitGroup
	var mapErr error

	for i, task := range j.Tasks {
		wg.Add(1)
		go func(t Task, idx int) {
			defer wg.Done()

			worker := j.Workers[idx%len(j.Workers)]
			j.logger.Printf("[MASTER] Assigning %s to %s\n", t.Path, worker)

			kvs, err := callWorkerMap(worker, j.Pattern, task.Path)
			if err != nil {
				j.logger.Printf("Error mapping %s: %v\n", t.Path, err)
				mapErr = err
				return
			}
			j.logger.Printf("Completed mapping %s. %d matches found\n", t.Path, len(kvs))

			mu.Lock()
			for _, kv := range kvs {
				j.logger.Printf("Intermediate Emit -> Key: %s\n", kv.Key)
				intermediates[kv.Key] = append(intermediates[kv.Key], kv.Value)
			}
			mu.Unlock()
		}(task, i)
	}

	wg.Wait()
	if mapErr != nil {
		return nil, mapErr
	}

	// Reduce Phase
	j.logger.Println("Starting Reduce Phase")
	results := make(map[string]string)

	fmt.Print(intermediates)
	for key, vals := range intermediates {
		j.logger.Printf("Reducing key: %s\n", key)
		out, err := j.Reducer.Reduce(key, vals)
		if err != nil {
			j.logger.Printf("Error reducing key %s: %v\n", key, err)
			return nil, err
		}
		results[key] = out
		j.logger.Printf("Reduced result: %s -> %s\n", key, out)
	}

	j.logger.Println("MapReduce Job Completed Successfully")
	return results, nil
}
