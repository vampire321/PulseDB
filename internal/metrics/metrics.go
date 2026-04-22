package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var(
	//this is for HTTP layer(RED method)
	HTTPRequestTotal = promauto.NewCounterVec(  //measure throughout and success rate .
		prometheus.CounterOpts{
			Name : "pulsedb_http_requests_total",
			Help : "Total HTTP requests by methos,path and status code",
		},[]string{"method","path","status_code"},
	)


	HttpRequestDuration= promauto.NewHistogramVec( //to measure latency (how long the user stays "waiting for an Api response")
		prometheus.HistogramOpts{
			Name : "pulsedb_http_request_duration_duration",
			Help : "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method","path"},
	)

//Business metrics

	ActiveMonitors = promauto.NewGauge( //to measure the Capacity/load
		prometheus.GaugeOpts{
			Name: "pulsedb_active_monitors",
			Help:"Number of active monitors",
		},
	)

	ChecksTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "pulsedb_checks_total",
            Help: "Total health checks executed",
        },
        []string{"result"}, // "success", "error", "timeout"
    )
	CheckDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "pulsedb_check_duration_seconds",
            Help:    "Health check probe duration in seconds",
            Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"monitor_name"},
    )
	  WorkerQueueDepth = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "pulsedb_worker_queue_depth",
            Help: "Number of monitors waiting in the worker queue",
        },
    )
)