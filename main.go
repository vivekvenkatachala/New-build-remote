package main

import (
	"log"
	"net/http"
	"start_build/controller"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	 
	port := "5000"

	r.HandleFunc("/start_build",controller.StartBuild)

	log.Printf("Connected to http://localhost:%s/ for remote_build backend", port)

	log.Fatal(http.ListenAndServe(":"+port, r))
}
	
	
