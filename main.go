package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sample-app/env"
	"sample-app/metrics"
	"sample-app/routes"
)

func main() {
	e := env.GetEnvironment()
	r := mux.NewRouter()
	metrics.RegisterInstrumentation(r)
	routes.RegisterHealthcheck(r)
	log.Printf("Starting server on port %s\n", e.HttpPort)
	log.Fatal(http.ListenAndServe(":" + e.HttpPort, r))
}
