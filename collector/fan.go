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

package collector

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const fanSubsystem = "fan"

type fanCollector struct {
	fanRPM     *prometheus.Desc
	fanPWMMode *prometheus.Desc
}

func init() {
	registerCollector("fan", defaultEnabled, NewFanCollector)
}

// NewFanCollector returns a new Collector exposing CPU temperature metrics.
func NewFanCollector() (Collector, error) {
	fc := &fanCollector{
		fanRPM: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, fanSubsystem, "rpm"),
			"Fan speed in RPM for Pi5",
			[]string{"hwmon"}, nil,
		),
		fanPWMMode: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, fanSubsystem, "pwm_level"),
			"Fan PWM level (for Pi 4 with PoE module)",
			[]string{"hwmon"}, nil,
		),
	}
	return fc, nil
}

// Update implements the Collector interface.
func (c *fanCollector) Update(ch chan<- prometheus.Metric) error {
	version := c.getPiVersion()

	if version == "5" {
		c.updateRPM(ch)
	} else {
		c.updatePWMMode(ch)
	}

	return nil
}

func (c *fanCollector) updateRPM(ch chan<- prometheus.Metric) error {
	// Get all the hwmons from /sys/devices/platform/cooling_fan/hwmon/
	hwmons, err := filepath.Glob("/sys/devices/platform/cooling_fan/hwmon/hwmon[0-9]*")
	if err != nil {
		return err
	}

	for _, hwmon := range hwmons {
		// Get the RPM string from /sys/devices/platform/cooling_fan/hwmon/hwmon*/fan1_input
		// and convert it to a float64 value.
		b, err := os.ReadFile(hwmon + "/fan1_input")
		if err != nil {
			return err
		}
		freq, err := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
		if err != nil {
			return err
		}

		// Export the metric.
		ch <- prometheus.MustNewConstMetric(
			c.fanRPM,
			prometheus.GaugeValue,
			freq,
			hwmon,
		)
	}

	return nil
}

func (c *fanCollector) updatePWMMode(ch chan<- prometheus.Metric) error {
	// Get all the hwmons from /sys/devices/platform/pwm-fan/hwmon/hwmon*
	hwmons, err := filepath.Glob("/sys/devices/platform/pwm-fan/hwmon/hwmon[0-9]*")
	if err != nil {
		return err
	}

	for _, hwmon := range hwmons {
		// Get the RPM string from /sys/devices/platform/pwm-fan/hwmon/hwmon*/pwm1
		// and convert it to a float64 value.
		b, err := os.ReadFile(hwmon + "/pwm1")
		if err != nil {
			return err
		}
		freq, err := strconv.ParseFloat(string(bytes.TrimSpace(b)), 64)
		if err != nil {
			return err
		}

		// Export the metric.
		ch <- prometheus.MustNewConstMetric(
			c.fanPWMMode,
			prometheus.GaugeValue,
			freq,
			hwmon,
		)
	}

	return nil
}

func (e *fanCollector) getPiVersion() string {
	fi, err := os.Open("/proc/cpuinfo")
	if err != nil {
		log.Fatal(err)
	}

	defer fi.Close()

	// Compile the regex pattern
	r := regexp.MustCompile(`^Model\s*:\sRaspberry Pi (\d).*$`)

	// make a read buffer
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := r.FindStringSubmatch(line); matches != nil {
			// matches[1] contains the first captured group
			return matches[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return "unknown"
}
