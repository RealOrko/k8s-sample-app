package routes

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sample-app/env"
	"sample-app/formatters"
	"time"
)

// *** Registration ***

func RegisterHealthcheck(r *mux.Router){
	r.HandleFunc("/test", handleHealthCheck)
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
		fmt.Fprintf(w, formatters.ToJSON(healthCheckResponse{
			Message:  "Not feeling great!",
			DateTime: time.Now(),
		}))
		log.Println("Warning! Health check is failing!")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, formatters.ToJSON(healthCheckResponse{
		Message:  "Feeling good!",
		DateTime: time.Now(),
	}))
	log.Println("Health check passed!")
}
