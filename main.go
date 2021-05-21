package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Environment struct {
	httpPort string
	failHealthCheck bool
}

func getEnvironmentKeyStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvironmentKeyBool(key, fallback string) bool {
	if value, ok := os.LookupEnv(key); ok {
		castValue, err := strconv.ParseBool(value)
		if err == nil {
			return castValue
		} else {
			log.Fatal(err)
		}
	}
	defaultValue, err := strconv.ParseBool(fallback)
	if err == nil {
		return defaultValue
	} else {
		log.Fatal(err)
	}
	panic(fmt.Sprintf("Cannot read environment variable '%s'", key))
}

func getEnvironment() Environment {
	return Environment{
		httpPort:        getEnvironmentKeyStr("PORT", "8000"),
		failHealthCheck: getEnvironmentKeyBool("FAIL_HEALTH_CHECK", "false"),
	}
}

func toJSON(i interface{}) string {
	res, err := json.Marshal(i)
	if err != nil {
		log.Fatal("Failed to convert to json ... ")
		panic("Exiting ... ")
	}
	return string(res)
}

type HealthCheckResponse struct {
	Message  string
	DateTime time.Time
}

var HandleHealthCheck = func(w http.ResponseWriter, r *http.Request) {
	e := getEnvironment()

	if e.failHealthCheck {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, toJSON(HealthCheckResponse{
			Message:  "Not feeling great!",
			DateTime: time.Now(),
		}))
		log.Println("Warning! Health check is failing!")
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, toJSON(HealthCheckResponse{
		Message:  "Feeling good!",
		DateTime: time.Now(),
	}))
	log.Println("Health check passed!")
}

func main() {
	e := getEnvironment()
	r := mux.NewRouter()
	r.HandleFunc("/api/health-check", HandleHealthCheck)
	log.Printf("Starting server on port %s\n", e.httpPort)
	log.Fatal(http.ListenAndServe(":" + e.httpPort, r))
}
