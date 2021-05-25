package routes

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sample-app/env"
	"time"
)

// *** Registration ***

func RegisterHealthcheck(r *mux.Router){
	r.HandleFunc("/test", handleHealthCheck)
}


// *** JSON Formatter ***

func toJSON(i interface{}) string {
	res, err := json.Marshal(i)
	if err != nil {
		log.Fatal("Failed to convert to json ... ")
		panic("Exiting ... ")
	}
	return string(res)
}

// *** Health Checks ***

type healthCheckResponse struct {
	Message  string
	DateTime time.Time
}

var handleHealthCheck = func(w http.ResponseWriter, r *http.Request) {
	e := env.GetEnvironment()

	if e.FailHealthCheck {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, toJSON(healthCheckResponse{
			Message:  "Not feeling great!",
			DateTime: time.Now(),
		}))
		log.Println("Warning! Health check is failing!")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, toJSON(healthCheckResponse{
		Message:  "Feeling good!",
		DateTime: time.Now(),
	}))
	log.Println("Health check passed!")
}
