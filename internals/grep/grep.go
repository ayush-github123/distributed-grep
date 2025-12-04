package grep

import (
	"bufio"
	"distGrep/internals/mapreduce"
	"fmt"
	"os"
	"regexp"
)

type GrepMapper struct {
	Pattern *regexp.Regexp
}

func NewGrepMapper(pattern string) (*GrepMapper, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return &GrepMapper{
		Pattern: re,
	}, nil
}

func (gm *GrepMapper) Map(task mapreduce.Task) ([]mapreduce.KeyValue, error) {

	file, err := os.Open(task.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var kvs []mapreduce.KeyValue
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		if gm.Pattern.MatchString(text) {
			kv := mapreduce.KeyValue{
				Key:   fmt.Sprintf("%s:%d", task.Path, lineNum),
				Value: text,
			}
			kvs = append(kvs, kv)
		}
	}
	return kvs, scanner.Err()
}


type GrepReducer struct {}


func (gr *GrepReducer) Reduce(key string, values []string) (string, error) {
	return values[0], nil
}

