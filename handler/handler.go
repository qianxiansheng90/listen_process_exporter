/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-04-30 16:38:47
 * @Description: desc
 */
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"listen_process_exporter/exporter"
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
