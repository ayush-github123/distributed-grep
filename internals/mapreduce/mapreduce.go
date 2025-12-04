package mapreduce

import (
	"log"
	"os"
	"sync"
)

type KeyValue struct {
	Key   string
	Value string
}

type Mapper interface {
	Map(task Task) ([]KeyValue, error)
}

type Reducer interface {
	Reduce(key string, values []string) (string, error)
}

type Task struct {
	Path string
}

type Job struct {
	Mapper  Mapper
	Pattern string
	Reducer Reducer
	Tasks   []Task
	Workers	[]string
	logger  *log.Logger
}

func NewJob(pattern string, mapper Mapper, reducer Reducer, tasks []Task, workers []string) *Job {
	logFile, err := os.OpenFile("distgrep.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error creating log file: %v", err)
	}

	logger := log.New(logFile, "[DistGrep] ", log.Ldate|log.Ltime|log.Lshortfile)

	return &Job{
		Mapper:  mapper,
		Pattern: pattern,
		Reducer: reducer,
		Tasks:   tasks,
		Workers: workers,
		logger:  logger,
	}
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

			worker := j.Workers[idx % len(j.Workers)]
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
