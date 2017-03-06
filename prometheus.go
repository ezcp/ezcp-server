package main

import (
	"log"
	"net/http"

	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	routesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "route_hits_counts",
		Help: "Route hit counters",
	}, []string{"host", "route", "status"})

	panicsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "route_panics_counts",
		Help: "Route panic counters",
	}, []string{"host", "route"})

	statusCodes = []int{200, 201, 401, 405, 500, 0} // 0 for unknown
)

type wrappedResponseWriter struct {
	counters map[int]prometheus.Counter
	res      http.ResponseWriter
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.res.Header()
}
func (w *wrappedResponseWriter) WriteHeader(status int) {
	counter, ok := w.counters[status]
	if ok {
		counter.Inc()
	} else {
		w.counters[0].Inc()
	}
	w.res.WriteHeader(status)
}
func (w *wrappedResponseWriter) Write(bytes []byte) (int, error) {
	return w.res.Write(bytes)
}

// hitCounter adds a counter to this route, which monitors hits and response codes
func hitCounter(host string, route string, f func(res http.ResponseWriter, req *http.Request)) func(res http.ResponseWriter, req *http.Request) {
	hitCounters := make(map[int]prometheus.Counter)
	for _, status := range statusCodes {
		hitCounters[status] = routesCounter.WithLabelValues(host, route, strconv.Itoa(status))
	}
	panicCounter := panicsCounter.WithLabelValues(host, route)
	return func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Print("recovered error ", err)
				panicCounter.Inc()
			}
		}()
		f(&wrappedResponseWriter{hitCounters, res}, req)
	}
}

func registerPrometheus() {
	prometheus.MustRegister(routesCounter)
	prometheus.MustRegister(panicsCounter)
}
