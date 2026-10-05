// Copyright 2019 Lukas Malkmus
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	promcollectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/sierrasoftworks/humane-errors-go"
	log "github.com/sirupsen/logrus"

	"github.com/spechtlabs/rpi_exporter/collector"
)

// A wrapper around http.Handler to handle filtering.
// Caches already used filter combinations.
// Create a new handler using newHandler().
type handler struct {
	unfilteredHandler http.Handler
	// There are only four collectors in this program, so that's fifteen
	// combinations at most. Concurrent scrapes share it, hence the mutex.
	filteredHandlers        map[string]http.Handler
	exporterMetricsRegistry *prometheus.Registry
	mu                      sync.Mutex
	includeExporterMetrics  bool
}

func newHandler(includeExporterMetrics bool) *handler {
	h := &handler{
		filteredHandlers:       make(map[string]http.Handler),
		includeExporterMetrics: includeExporterMetrics,
	}

	// Add default collectors, if they aren't disabled.
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry = prometheus.NewRegistry()
		h.exporterMetricsRegistry.MustRegister(
			promcollectors.NewProcessCollector(promcollectors.ProcessCollectorOpts{}),
			promcollectors.NewGoCollector(),
		)
	}

	// Create the unfiltered default handler.
	unfilteredHandler, err := h.filteredHandler()
	if err != nil {
		panic("Couldn't create metrics handler: " + err.Display())
	}

	h.unfilteredHandler = unfilteredHandler
	return h
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the filters from the query.
	filters := r.URL.Query()["collect[]"]
	// Sort filters, and drop repeated ones, to allow caching of filtered
	// handlers: collect[]=cpu&collect[]=cpu is the same handler as
	// collect[]=cpu, not a new cache entry.
	sort.Strings(filters)
	filters = slices.Compact(filters)
	log.Debugln("collect query:", filters)

	// Use the unfiltered handler if no filters were given.
	if len(filters) == 0 {
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}

	// Create a filtered handler.
	filteredHandler, err := h.filteredHandler(filters...)
	if err != nil {
		log.Errorln("Couldn't create filtered handler:", err.Display())
		// The message names the collect[] values the request sent; plain text
		// keeps a browser from rendering them.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "Couldn't create filtered metrics handler: "+err.Display())
		return
	}

	filteredHandler.ServeHTTP(w, r)
}

func (h *handler) filteredHandler(filters ...string) (http.Handler, humane.Error) {
	// Do not recreate unfiltered handler if it already exists.
	if len(filters) == 0 && h.unfilteredHandler != nil {
		return h.unfilteredHandler, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if there is a handler for this combination of filters already.
	filtersStr := strings.Join(filters, ",")
	handler := h.filteredHandlers[filtersStr]
	if handler != nil {
		return handler, nil
	}

	// Create a new Raspberry Pi collector.
	rpiColl, err := collector.New(filters...)
	if err != nil {
		return nil, err
	}

	// Create a new prometheus registry and register the Raspberry Pi collector
	// on it.
	reg := prometheus.NewRegistry()
	if err := reg.Register(rpiColl); err != nil {
		return nil, humane.Wrap(err, "cannot register the collectors", "this is a bug in rpi_exporter; please report it at https://github.com/SpechtLabs/rpi_exporter/issues")
	}

	// Delegate http serving to Prometheus client library, which will call
	// collector.Collect.
	if h.includeExporterMetrics {
		handler = promhttp.HandlerFor(
			prometheus.Gatherers{h.exporterMetricsRegistry, reg},
			promhttp.HandlerOpts{
				ErrorLog:      log.StandardLogger(),
				ErrorHandling: promhttp.HTTPErrorOnError,
				Registry:      h.exporterMetricsRegistry,
			})
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	} else {
		handler = promhttp.HandlerFor(reg,
			promhttp.HandlerOpts{
				ErrorLog:      log.StandardLogger(),
				ErrorHandling: promhttp.HTTPErrorOnError,
			})
	}

	// Store handler in cache if it isn't unfiltered.
	if len(filters) > 0 {
		h.filteredHandlers[filtersStr] = handler
	}
	return handler, nil
}

// HealthCheckHandler answers the health check path with a static JSON
// document: if it answers at all, the exporter is alive.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `{"alive": true}`)
}

// indexHandler serves the landing page, which links the metrics and the
// health check.
func indexHandler(metricsPath, healthPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Raspberry Pi Exporter</title></head>
			<body>
			<h1>Raspberry Pi Exporter</h1>
			<p><a href="` + metricsPath + `">Metrics</a></p>
			<p><a href="` + healthPath + `">Exporter health</a></p>
			</body>
			</html>`))
	}
}

func main() {
	// Command line flags.
	var (
		webListenAddress          = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":9243").String()
		webMetricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		webHealthPath             = kingpin.Flag("web.healthcheck-path", "Path under which the exporter exposes its status.").Default("/health").String()
		webDisableExporterMetrics = kingpin.Flag("web.disable-exporter-metrics", "Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).").Bool()
	)

	// Setup the command line flags and commands.
	collector.RegisterFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("rpi_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	// Print build context and version.
	log.Info("Starting rpi_exporter", version.Info())
	log.Info("Build context", version.BuildContext())

	// Setup router and handlers.
	mux := http.NewServeMux()
	mux.Handle(*webMetricsPath, newHandler(!*webDisableExporterMetrics))
	mux.HandleFunc(*webHealthPath, HealthCheckHandler)
	mux.HandleFunc("/", indexHandler(*webMetricsPath, *webHealthPath))

	// Setup webserver.
	srv := &http.Server{
		Addr:         *webListenAddress,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Listen for termination signals.
	term := make(chan os.Signal, 1)
	defer close(term)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)

	// Run webserver in a separate go-routine.
	log.Info("Listening on ", *webListenAddress)
	webErr := make(chan error)
	defer close(webErr)
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			webErr <- err
		}
	}()

	// Wait for a termination signal and shut down gracefully, but wait no
	// longer than 5 seconds before halting.
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	select {
	case <-term:
		log.Warn("Received SIGTERM, exiting gracefully...")
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error(err)
		}
	case err := <-webErr:
		log.Error("Error starting web server, exiting gracefully:", err)
	}
	log.Info("See you next time!")
}
