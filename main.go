package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
)



type Match struct{
	File string
	Line int
	Text string
}



func grepFile(pattern, path string) ([]Match, error) {
	regPattern, err := regexp.Compile(pattern)
	if err!=nil{
		return nil, fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	file, err := os.Open(path)
	if err!=nil{
		return nil, fmt.Errorf("error opening file %w", err)
	}
	defer file.Close()

	var matches []Match

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan(){
		lineNum++

		text := scanner.Text()

		if regPattern.MatchString(text){
			matches = append(matches, Match{
				File: path,
				Line: lineNum,
				Text: text,
			})
		}
	}

	if err:=scanner.Err(); err!=nil{
		return nil, fmt.Errorf("error reading file %q: %w", path, err)
	}

	return matches, nil

}



func grepFiles(pattern string, paths []string) ([]Match, error) {
	var all []Match

	for _, p := range paths {
		matches, err := grepFile(pattern, p)
		if err != nil {
			// You can decide later to log and continue instead of returning.
			return nil, err
		}
		all = append(all, matches...)
	}

	return all, nil
}




func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <pattern> <file1> [file2...]\n", os.Args[0])
		os.Exit(1)
	}

	pattern := os.Args[1]
	files := os.Args[2:]

	matches, err := grepFiles(pattern, files)
	if err != nil {
		log.Fatalf("grep error: %v", err)
	}

	for _, m := range matches {
		fmt.Printf("%s:%d: %s\n", m.File, m.Line, m.Text)
	}
}