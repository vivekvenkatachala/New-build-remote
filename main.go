package main

import (
	"log"
	"net/http"
	"start_build/controller"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/start_build",controller.StartBuild)

	log.Fatal(http.ListenAndServe(":5000", r))
}
	
	
