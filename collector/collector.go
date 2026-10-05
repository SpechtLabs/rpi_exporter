// Copyright 2015 The Prometheus Authors
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

package collector

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sierrasoftworks/humane-errors-go"
	log "github.com/sirupsen/logrus"
)

// Namespace defines the common namespace to be used by all metrics.
const namespace = "rpi"

const defaultEnabled = true

var (
	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"rpi_exporter: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"rpi_exporter: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

var (
	factories      = make(map[string]func() (Collector, error))
	collectorState = make(map[string]*bool)
)

// RegisterFlags defines the collectors' command line flags on app: one
// --collector.<name> flag per collector, and the flags the collectors read
// their settings from. Call it once, before app.Parse; New builds only the
// collectors whose flags are set.
func RegisterFlags(app *kingpin.Application) {
	vcgencmd = app.Flag("vcgencmd", "vcgencmd including path.").Default("/opt/vc/bin/vcgencmd").String()
	textFileDirectory = app.Flag("collector.textfile.directory", "Directory to read text files with metrics from.").Default("").String()

	registerCollector(app, "cpu", defaultEnabled, NewCPUCollector)
	registerCollector(app, "fan", defaultEnabled, NewFanCollector)
	registerCollector(app, "gpu", defaultEnabled, NewGPUCollector)
	registerCollector(app, "textfile", defaultEnabled, NewTextFileCollector)
}

// registerCollector adds the --collector.<name> flag that enables a collector,
// and the factory New builds it with.
func registerCollector(app *kingpin.Application, collector string, isDefaultEnabled bool, factory func() (Collector, error)) {
	// Get the default state as a string for the help flag.
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	// Create the flags for the givec RPiCollector.
	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := app.Flag(flagName, flagHelp).Default(defaultValue).Bool()
	collectorState[collector] = flag

	factories[collector] = factory
}

// Collector is the interface a collector has to implement.
type Collector interface {
	// Get new metrics and expose them via prometheus registry.
	Update(ch chan<- prometheus.Metric) error
}

// RPiCollector implements the prometheus.Collector interface.
type RPiCollector struct {
	collectors map[string]Collector
}

// New creates a new Raspberry Pi collector with the enabled collectors, or
// only those of them that filters names.
func New(filters ...string) (*RPiCollector, humane.Error) {
	// Build the map of requested/filtered collectors.
	f := make(map[string]bool)
	for _, filter := range filters {
		enabled, exist := collectorState[filter]
		if !exist {
			return nil, humane.New(fmt.Sprintf("unknown collector %q", filter), "collect[] takes one of: "+strings.Join(slices.Sorted(maps.Keys(collectorState)), ", "))
		}
		if !*enabled {
			return nil, humane.New(fmt.Sprintf("collector %q is disabled", filter), fmt.Sprintf("start rpi_exporter with --collector.%s to enable it", filter))
		}
		f[filter] = true
	}

	// Get the requested collectors.
	collectors := make(map[string]Collector)
	for key, enabled := range collectorState {
		if *enabled {
			collector, err := factories[key]()
			if err != nil {
				return nil, humane.Wrap(err, fmt.Sprintf("cannot create the %s collector", key), fmt.Sprintf("start rpi_exporter with --no-collector.%s to run without it", key))
			}
			if len(f) == 0 || f[key] {
				collectors[key] = collector
			}
		}
	}
	return &RPiCollector{collectors}, nil
}

// Describe implements the prometheus.Collector interface.
func (c RPiCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (c RPiCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.collectors))
	for name, c := range c.collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	// Update the collector and meassure its execution time.
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	// Log the execution status and set the appropriate success value.
	if err != nil {
		log.Errorf("%s collector failed after %fs: %s", name, duration.Seconds(), err)
		success = 0
	} else {
		log.Debugf("%s collector succeeded after %fs", name, duration.Seconds())
		success = 1
	}

	// Record execution time and success value.
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
