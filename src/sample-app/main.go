package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type HealthCheckResponse struct {
	Message  string
	DateTime time.Time
}

var HandleHealthCheck = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	response := HealthCheckResponse{
		Message:  "Feeling good!",
		DateTime: time.Now(),
	}
	res, _ := json.Marshal(response)
	fmt.Fprintf(w, string(res))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/health-check", HandleHealthCheck)
	log.Fatal(http.ListenAndServe(":8000", r))
}
