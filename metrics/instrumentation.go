package metrics

//#include <time.h>
import "C"

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"runtime"
	"sample-app/env"
	"strconv"
	"strings"
	"time"
)

// *** Registration ***

func RegisterInstrumentation(r *mux.Router){
	r.Use(prometheusMiddleware)
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
	prometheus.Register(cpuUsagePercent)
	prometheus.Register(memAllocatedTotal)
	r.Path("/metrics").Handler(promhttp.Handler())

}

// *** CPU Measurements ***

var startTime = time.Now()
var startTicks = C.clock()

func getCpuUsagePercent() float64 {
	clockSeconds := float64(C.clock()-startTicks) / float64(C.CLOCKS_PER_SEC)
	realSeconds := time.Since(startTime).Seconds()
	return clockSeconds / realSeconds * 100
}

// *** Prometheus ***

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
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

var memAllocatedTotal = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "memory_allocated_mb",
		Help: "Status of HTTP response",
	})

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

var cpuUsagePercent = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "cpu_usage_percent",
	Help: "The usage of the CPU in %",
}, []string{"default"})

// *** Helpers ***

func randomDelay(){
	randomNumber := rand.Intn(env.GetEnvironment().RequestDelay)
	time.Sleep(time.Duration(randomNumber) * time.Second)
}

func bytesToMegabyte(b uint64) float64 {
	return float64(b / 1024 / 1024)
}

func getMemory() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return bytesToMegabyte(m.Alloc)
}

// *** Middleware ***

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))

		if ! strings.Contains(path, "test") {
			randomDelay()
		}

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		statusCode := rw.statusCode

		timer.ObserveDuration()
		memAllocatedTotal.Set(getMemory())
		totalRequests.WithLabelValues(path).Inc()
		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		cpuUsagePercent.WithLabelValues("default").Observe(getCpuUsagePercent())
	})
}