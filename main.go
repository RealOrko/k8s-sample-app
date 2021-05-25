package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Environment struct {
	httpPort        string
	failHealthCheck bool
	requestDelay    int
}

func getEnvironmentKeyStr(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvironmentKeyBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		castValue, err := strconv.ParseBool(value)
		if err == nil {
			return castValue
		} else {
			log.Fatal(err)
		}
	}
	return fallback
}

func getEnvironmentKeyInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		castValue, err := strconv.Atoi(value)
		if err == nil {
			return castValue
		} else {
			log.Fatal(err)
		}
	}
	return fallback
}

func getEnvironment() Environment {
	env := Environment{
		httpPort:        getEnvironmentKeyStr("PORT", "8000"),
		failHealthCheck: getEnvironmentKeyBool("FAIL_HEALTH_CHECK", false),
		requestDelay: 	 getEnvironmentKeyInt("REQUEST_DELAY", 5),
	}
	log.Printf("env:PORT: %s\n", env.httpPort)
	log.Printf("env:REQUEST_DELAY: %v\n", env.requestDelay)
	log.Printf("env:FAIL_HEALTH_CHECK: %v\n", env.failHealthCheck)

	return env
}

func toJSON(i interface{}) string {
	res, err := json.Marshal(i)
	if err != nil {
		log.Fatal("Failed to convert to json ... ")
		panic("Exiting ... ")
	}
	return string(res)
}

// *** Health Checks ***

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

// *** Prometheus ***

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "debug_app_http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "debug_app_response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "debug_app_http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

// *** Random Delays ***

func randomDelay(){
	randomNumber := rand.Intn(getEnvironment().requestDelay)
	time.Sleep(time.Duration(randomNumber) * time.Second)
}

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		if ! strings.Contains(path, "test") {
			randomDelay()
		}
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequests.WithLabelValues(path).Inc()
		timer.ObserveDuration()
	})
}

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}

// *** Main ***

func main() {
	e := getEnvironment()
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)

	r.HandleFunc("/test", HandleHealthCheck)
	r.Path("/metrics").Handler(promhttp.Handler())

	log.Printf("Starting server on port %s\n", e.httpPort)
	log.Fatal(http.ListenAndServe(":" + e.httpPort, r))
}
