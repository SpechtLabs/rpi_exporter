package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFiles creates files, keyed by their path under root, with their
// contents.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, contents := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	}
}

// update runs a collector's Update and returns what it sent, in the text
// exposition format.
func update(t *testing.T, c Collector) (string, error) {
	t.Helper()
	ch := make(chan prometheus.Metric, 64)
	err := c.Update(ch)
	close(ch)

	reg := prometheus.NewPedanticRegistry()
	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}
	require.NoError(t, reg.Register(collected(metrics)))
	families, gatherErr := reg.Gather()
	require.NoError(t, gatherErr)

	var out strings.Builder
	for _, f := range families {
		_, encodeErr := expfmt.MetricFamilyToText(&out, f)
		require.NoError(t, encodeErr)
	}
	return out.String(), err
}

// collected replays metrics an Update sent, as an unchecked collector.
type collected []prometheus.Metric

func (collected) Describe(chan<- *prometheus.Desc) {}

func (c collected) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c {
		ch <- m
	}
}

func TestCPUCollector(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		want    []string
		wantErr bool
	}{
		{
			name: "temperature and frequencies",
			files: map[string]string{
				"class/thermal/thermal_zone0/temp":                 "48312\n",
				"devices/system/cpu/cpu0/cpufreq/scaling_cur_freq": "1500000\n",
				"devices/system/cpu/cpu1/cpufreq/scaling_cur_freq": "600000\n",
			},
			want: []string{
				"rpi_cpu_temperature_celsius 48.312",
				`rpi_cpu_frequency_hertz{cpu="0"} 1.5e+06`,
				`rpi_cpu_frequency_hertz{cpu="1"} 600000`,
			},
		},
		{
			name:    "no thermal zone",
			wantErr: true,
		},
		{
			name:    "unreadable temperature",
			files:   map[string]string{"class/thermal/thermal_zone0/temp": "hot\n"},
			wantErr: true,
		},
		{
			name: "cpu without cpufreq",
			files: map[string]string{
				"class/thermal/thermal_zone0/temp": "48312\n",
				"devices/system/cpu/cpu0/online":   "1\n",
			},
			wantErr: true,
		},
		{
			name: "unreadable frequency",
			files: map[string]string{
				"class/thermal/thermal_zone0/temp":                 "48312\n",
				"devices/system/cpu/cpu0/cpufreq/scaling_cur_freq": "fast\n",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCPUCollector()
			require.NoError(t, err)
			c.(*cpuCollector).sysfs = t.TempDir()
			writeFiles(t, c.(*cpuCollector).sysfs, tt.files)

			got, err := update(t, c)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			for _, want := range tt.want {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestFanCollector(t *testing.T) {
	pi5 := "Model\t\t: Raspberry Pi 5 Model B Rev 1.0\n"
	pi4 := "Model\t\t: Raspberry Pi 4 Model B Rev 1.4\n"

	tests := []struct {
		name    string
		cpuinfo *string
		files   map[string]string
		want    []string
		wantErr bool
	}{
		{
			name:    "Pi 5 fan speed",
			cpuinfo: &pi5,
			files:   map[string]string{"devices/platform/cooling_fan/hwmon/hwmon2/fan1_input": "2950\n"},
			want:    []string{`rpi_fan_rpm{hwmon="SYSFS/devices/platform/cooling_fan/hwmon/hwmon2"} 2950`},
		},
		{
			name:    "Pi 4 PoE fan level",
			cpuinfo: &pi4,
			files:   map[string]string{"devices/platform/pwm-fan/hwmon/hwmon1/pwm1": "75\n"},
			want:    []string{`rpi_fan_pwm_level{hwmon="SYSFS/devices/platform/pwm-fan/hwmon/hwmon1"} 75`},
		},
		{
			name:    "Pi 4 without a fan",
			cpuinfo: &pi4,
		},
		{
			name:    "no cpuinfo",
			wantErr: true,
		},
		{
			name:    "Pi 5 fan speed unreadable",
			cpuinfo: &pi5,
			files:   map[string]string{"devices/platform/cooling_fan/hwmon/hwmon2/fan1_input": "spinning\n"},
			wantErr: true,
		},
		{
			name:    "Pi 5 fan without fan1_input",
			cpuinfo: &pi5,
			files:   map[string]string{"devices/platform/cooling_fan/hwmon/hwmon2/name": "pwmfan\n"},
			wantErr: true,
		},
		{
			name:    "Pi 4 fan level unreadable",
			cpuinfo: &pi4,
			files:   map[string]string{"devices/platform/pwm-fan/hwmon/hwmon1/pwm1": "high\n"},
			wantErr: true,
		},
		{
			name:    "Pi 4 fan without pwm1",
			cpuinfo: &pi4,
			files:   map[string]string{"devices/platform/pwm-fan/hwmon/hwmon1/name": "pwmfan\n"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewFanCollector()
			require.NoError(t, err)
			fan := c.(*fanCollector)
			fan.sysfs, fan.procfs = t.TempDir(), t.TempDir()
			writeFiles(t, fan.sysfs, tt.files)
			if tt.cpuinfo != nil {
				writeFiles(t, fan.procfs, map[string]string{"cpuinfo": *tt.cpuinfo})
			}

			got, err := update(t, c)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			for _, want := range tt.want {
				assert.Contains(t, got, strings.ReplaceAll(want, "SYSFS", fan.sysfs))
			}
		})
	}
}
