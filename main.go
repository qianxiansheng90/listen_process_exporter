/*
 * @Description: desc
 */
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"listen_process_exporter/comm"
	"listen_process_exporter/exporter"
	"listen_process_exporter/handler"
	"listen_process_exporter/listen_process"
)

var (
	listenAddress = flag.String("web.listen-address", ":9911",
		"Addresses on which to expose metrics and web")
	metricsPath = flag.String("web.telemetry-path", "/metrics",
		"Path under which to expose metrics")
	collectChildProcess          = flag.Bool("collector.child", false, "Enable the collect child process (default: disable).")
	refreshListenProcessInterval = flag.Int("collector.refresh", 60, "Refresh listen process interval second (default: 60s).")
	collectListenPort            = flag.Int("collector.port", 3306, "Collect listen port (default: 3306).")
	debug                        = flag.Bool("collector.debug", false, "Enable debug mode.")
)

func main() {
	flag.Parse()
	if *refreshListenProcessInterval < 5 {
		if *refreshListenProcessInterval != listen_process.ForbidInterval {
			log.Printf("Error: collector.refresh too small (%ds)", *refreshListenProcessInterval)
			return
		}
	}
	if *debug {
		comm.SetDebug(*debug)
	}
	newExporter := exporter.NewExporter(*collectChildProcess, *collectListenPort)
	prometheus.MustRegister(newExporter)

	if err := listen_process.SetTickerInterval(*refreshListenProcessInterval); err != nil {
		log.Fatal(err)
		return
	}
	go listen_process.RefreshListenProcessGoroutine()

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/probe", handler.HandleProbe(*collectChildProcess))
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
