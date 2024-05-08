/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-04-30 16:38:47
 * @Description: desc
 */
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"listen_process_exporter/comm"
	"listen_process_exporter/exporter"
	"listen_process_exporter/listen_process"
)

func HandleProbe(collectChildProcess bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		target := params.Get("target")
		if target == "" {
			http.Error(w, "target is required", http.StatusBadRequest)
			return
		}
		listenPort, err := strconv.Atoi(target)
		if err != nil {
			http.Error(w, fmt.Sprintf("target[%s] must be number", target), http.StatusBadRequest)
			return
		}
		registry := prometheus.NewRegistry()

		registry.MustRegister(exporter.NewExporter(collectChildProcess, listenPort))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

/*
 *  @Description: refresh listen process through http
 */
func HandleRefreshListenProcess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if comm.Debug() {
			log.Printf("refresh listen process through http request %s", r.RemoteAddr)
		}
		listenProcess, err := listen_process.RefreshListenProcess(context.TODO())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			log.Printf("refresh listen process fail: %v", err)
			return
		}
		listenProcesslistJson, _ := json.Marshal(listenProcess)
		http.Error(w, string(listenProcesslistJson), http.StatusOK)
	}
}
