package collector

import (
	"errors"
	"maps"
	"slices"
	"testing"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseFlags registers the collectors' flags on an application of its own
// and parses args, as main does with the command line.
func parseFlags(t *testing.T, args ...string) {
	t.Helper()
	app := kingpin.New("rpi_exporter", "")
	RegisterFlags(app)
	_, err := app.Parse(args)
	require.NoError(t, err)
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		filters        []string
		wantCollectors []string
		wantErr        string
		wantAdvice     []string
	}{
		{
			name:           "every collector by default",
			wantCollectors: []string{"cpu", "fan", "gpu", "textfile"},
		},
		{
			name:           "only the enabled collectors",
			args:           []string{"--no-collector.gpu", "--no-collector.textfile"},
			wantCollectors: []string{"cpu", "fan"},
		},
		{
			name:           "only the filtered collectors",
			filters:        []string{"cpu", "gpu"},
			wantCollectors: []string{"cpu", "gpu"},
		},
		{
			name:       "unknown collector",
			filters:    []string{"disk"},
			wantErr:    `unknown collector "disk"`,
			wantAdvice: []string{"collect[] takes one of: cpu, fan, gpu, textfile"},
		},
		{
			name:       "disabled collector",
			args:       []string{"--no-collector.gpu"},
			filters:    []string{"gpu"},
			wantErr:    `collector "gpu" is disabled`,
			wantAdvice: []string{"start rpi_exporter with --collector.gpu to enable it"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parseFlags(t, tt.args...)

			got, err := New(tt.filters...)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				assert.Equal(t, tt.wantAdvice, err.Advice())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCollectors, slices.Sorted(maps.Keys(got.collectors)))
		})
	}
}

func TestNewFactoryError(t *testing.T) {
	parseFlags(t)
	failing := errors.New("no sysfs")
	t.Cleanup(func() { factories["cpu"] = NewCPUCollector })
	factories["cpu"] = func() (Collector, error) { return nil, failing }

	_, err := New()
	require.Error(t, err)
	assert.ErrorIs(t, err, failing)
	assert.Equal(t, "cannot create the cpu collector", err.Error())
	assert.Equal(t, []string{"start rpi_exporter with --no-collector.cpu to run without it"}, err.Advice())
}

// stubCollector reports one gauge, or fails.
type stubCollector struct {
	err error
}

func (s stubCollector) Update(ch chan<- prometheus.Metric) error {
	if s.err != nil {
		return s.err
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("rpi_stub", "A stub metric.", nil, nil),
		prometheus.GaugeValue, 1,
	)
	return nil
}

func TestCollect(t *testing.T) {
	c := RPiCollector{collectors: map[string]Collector{
		"ok":     stubCollector{},
		"failed": stubCollector{err: errors.New("no sysfs")},
	}}

	reg := prometheus.NewRegistry()
	require.NoError(t, reg.Register(c))
	families, err := reg.Gather()
	require.NoError(t, err)

	success := map[string]float64{}
	for _, f := range families {
		if f.GetName() != "rpi_scrape_collector_success" {
			continue
		}
		for _, m := range f.GetMetric() {
			success[m.GetLabel()[0].GetValue()] = m.GetGauge().GetValue()
		}
	}
	assert.Equal(t, map[string]float64{"ok": 1, "failed": 0}, success)
}
