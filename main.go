package main

//#include <time.h>
import "C"

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
	"runtime"
	"strconv"
	"strings"
	"time"
)

// *** Environment ***

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

// *** CPU Measurements ***

var startTime = time.Now()
var startTicks = C.clock()

func CpuUsagePercent() float64 {
	clockSeconds := float64(C.clock()-startTicks) / float64(C.CLOCKS_PER_SEC)
	realSeconds := time.Since(startTime).Seconds()
	return clockSeconds / realSeconds * 100
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

var cpuUsagePercent = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "debug_app_cpu_usage_percent",
	Help: "The usage of the CPU in %",
}, []string{"default"})

var memAllocatedTotal = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "debug_app_memory_allocated_mb",
	Help: "The total app memory allocated in Mb",
}, []string{"default"})

// *** Random Delays ***

func randomDelay(){
	randomNumber := rand.Intn(getEnvironment().requestDelay)
	time.Sleep(time.Duration(randomNumber) * time.Second)
}

// *** Helpers ***

func bToMb(b uint64) float64 {
	return float64(b / 1024 / 1024)
}

// *** Prometheus Middleware ***

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

		timer.ObserveDuration()
		totalRequests.WithLabelValues(path).Inc()
		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		cpuUsagePercent.WithLabelValues("default").Observe(CpuUsagePercent())

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		memAllocatedTotal.WithLabelValues("default").Observe(bToMb(m.TotalAlloc))
	})
}

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
	prometheus.Register(cpuUsagePercent)
	prometheus.Register(memAllocatedTotal)
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
