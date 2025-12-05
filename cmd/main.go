package main

import (
	"distGrep/cmd/master"
	"log"
	"net/http"
)

func main() {
	log.SetPrefix("[MASTER] ")

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	http.HandleFunc("/register", master.HandleRegister)
	http.HandleFunc("/run", master.HandleRunJob)

	log.Println("Master server running on :9000")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatalf("Master server error: %v", err)
	}
}
