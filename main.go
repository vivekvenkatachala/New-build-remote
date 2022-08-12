package main

import (
	"log"
	"net/http"
	"start_build/controller"

	"github.com/alecthomas/log4go"
	"github.com/gorilla/mux"
	"start_build/service_logs"
)

func main() {
	r := mux.NewRouter()
	 
	port := "5000"

	r.HandleFunc("/start_build",controller.StartBuild)

	log.Printf("Connected to http://localhost:%s/ for remote_build backend", port)

	servicelogs.Init()
	log4go.LoadConfiguration("log-config.xml")

	log.Fatal(http.ListenAndServe(":"+port, r))

}
	
	
