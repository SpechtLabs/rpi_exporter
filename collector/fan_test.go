package collector

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePiVersion(t *testing.T) {
	tests := []struct {
		name    string
		cpuinfo string
		want    string
	}{
		{
			name:    "Raspberry Pi 5",
			cpuinfo: "processor\t: 0\nRevision\t: d04170\nModel\t\t: Raspberry Pi 5 Model B Rev 1.0\n",
			want:    "5",
		},
		{
			name:    "Raspberry Pi 4",
			cpuinfo: "processor\t: 0\nModel\t\t: Raspberry Pi 4 Model B Rev 1.4\n",
			want:    "4",
		},
		{
			name:    "not a Raspberry Pi",
			cpuinfo: "processor\t: 0\nmodel name\t: Intel(R) Xeon(R)\n",
			want:    "unknown",
		},
		{
			name: "empty",
			want: "unknown",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePiVersion(strings.NewReader(tt.cpuinfo))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParsePiVersionReadError(t *testing.T) {
	// A line longer than bufio.Scanner's buffer makes the scan fail.
	_, err := parsePiVersion(strings.NewReader(strings.Repeat("x", 70*1024)))
	assert.Error(t, err)
}
