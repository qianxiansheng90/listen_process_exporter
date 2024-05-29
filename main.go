/*
 * @Description: desc
 */
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"listen_process_exporter/comm"
	"listen_process_exporter/exporter"
	"listen_process_exporter/handler"
	"listen_process_exporter/listen_process"
)

var (
	version       = flag.Bool("version", false, "Print version information.")
	listenAddress = flag.String("web.listen-address", ":9911",
		"Addresses on which to expose metrics and web")
	metricsPath = flag.String("web.telemetry-path", "/metrics",
		"Path under which to expose metrics")
	//collectChildProcess          = flag.Bool("collector.child", false, "Enable the collect child process (default: disable).")
	refreshListenProcessInterval = flag.Int("collector.refresh", 60, "Refresh listen process interval second (default: 60s).")
	collectListenPort            = flag.Int("collector.port", 3306, "Collect listen port (default: 3306).")
	debug                        = flag.Bool("collector.debug", false, "Enable debug mode.")
)

func main() {
	flag.Parse()
	if *version {
		fmt.Println(comm.Version)
		return
	}
	if *refreshListenProcessInterval < 5 {
		if *refreshListenProcessInterval != listen_process.ForbidInterval {
			log.Printf("Error: collector.refresh too small (%ds)", *refreshListenProcessInterval)
			return
		}
	}
	if *debug {
		comm.SetDebug(*debug)
	}

	handlerFunc := newHandler()
	http.Handle(*metricsPath, promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, handlerFunc))

	if err := listen_process.SetTickerInterval(*refreshListenProcessInterval); err != nil {
		log.Fatal(err)
		return
	}
	go listen_process.RefreshListenProcessGoroutine()

	//http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/probe", handler.HandleProbe(false))
	http.HandleFunc("/refresh_listen_process", handler.HandleRefreshListenProcess())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Mirth Channel Exporter</title></head>
             <body>
             <h1>Mirth Channel Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Printf("Listening on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func newHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		target := strconv.Itoa(*collectListenPort)
		q := r.URL.Query()
		if q.Has("target") {
			target = q.Get("target")
		}
		listenPort, err := strconv.Atoi(target)
		if err != nil {
			http.Error(w, fmt.Sprintf("target[%s] must be number", target), http.StatusBadRequest)
			return
		}
		registry := prometheus.NewRegistry()
		registry.MustRegister(exporter.NewExporter(false, listenPort))

		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			registry,
		}
		// Delegate http serving to Prometheus client library, which will call collector.Collect.
		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
