package main

import (
	"distGrep/internals/grep"
	"distGrep/internals/mapreduce"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: distgrep <pattern> <file1> [file2...]")
		return
	}

	pattern := os.Args[1]
	files := os.Args[2:]

	mapper, err := grep.NewGrepMapper(pattern)
	if err != nil {
		panic(err)
	}

	var tasks []mapreduce.Task

	for _, f := range files {
		tasks = append(tasks, mapreduce.Task{Path: f})
	}

	// job := mapreduce.NewJob(mapper, &grep.GrepReducer{}, tasks)
	// results, err := job.Execute()
	
	workers := []string{
		"http://localhost:8001",
		"http://localhost:8002",
	}
	
	job := mapreduce.NewJob(pattern, mapper, &grep.GrepReducer{}, tasks, workers)
	// job := mapreduce.Job{
	// 	Pattern:  pattern,
	// 	Tasks:    tasks,
	// 	Reducer:  &grep.GrepReducer{},
	// 	Workers:  workers,
	// }
	
	results, err := job.Execute()
	if err != nil {
		fmt.Print("Error in Execution")
		panic(err)
	}

	for k, v := range results {
		fmt.Println(k, v)
	}

	
	
}
