package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFileCollector(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{
			name: "metrics from .prom files",
			files: map[string]string{
				// The two series have different labels; each gets the other's
				// as an empty label.
				"backup.prom": "# TYPE backup_last_success gauge\nbackup_last_success{job=\"home\"} 1\nbackup_last_success{host=\"pi\"} 0\n",
				"notes.txt":   "not a metrics file\n",
			},
			want: []string{
				`backup_last_success{host="",job="home"} 1`,
				`backup_last_success{host="pi",job=""} 0`,
				`node_textfile_mtime_seconds{file="backup.prom"}`,
				"node_textfile_scrape_error 0",
			},
		},
		{
			name: "every metric type",
			files: map[string]string{
				"types.prom": "# TYPE c counter\nc 1\n# TYPE u untyped\nu 2\n" +
					"# TYPE s summary\ns{quantile=\"0.5\"} 3\ns_sum 3\ns_count 1\n" +
					"# TYPE h histogram\nh_bucket{le=\"1\"} 1\nh_bucket{le=\"+Inf\"} 1\nh_sum 0.5\nh_count 1\n",
			},
			want: []string{"c 1", "u 2", `s{quantile="0.5"} 3`, `h_bucket{le="1"} 1`, "node_textfile_scrape_error 0"},
		},
		{
			name:  "unparsable file",
			files: map[string]string{"broken.prom": "this is not { the text format\n"},
			want:  []string{"node_textfile_scrape_error 1"},
		},
		{
			name:  "client-side timestamps",
			files: map[string]string{"stamped.prom": "stamped 1 1700000000000\n"},
			want:  []string{"node_textfile_scrape_error 1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, tt.files)

			got, err := update(t, &textFileCollector{path: dir})
			require.NoError(t, err)
			for _, want := range tt.want {
				assert.Contains(t, got, want)
			}
		})
	}
}
