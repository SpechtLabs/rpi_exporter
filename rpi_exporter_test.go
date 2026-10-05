package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spechtlabs/rpi_exporter/collector"
)

func TestHandler(t *testing.T) {
	app := kingpin.New("rpi_exporter", "")
	collector.RegisterFlags(app)
	_, err := app.Parse(nil)
	require.NoError(t, err)

	tests := []struct {
		name            string
		query           string
		wantStatus      int
		wantContentType string
		wantBody        []string
	}{
		{
			name:       "every collector",
			wantStatus: http.StatusOK,
			wantBody:   []string{`rpi_scrape_collector_success{collector="cpu"}`, "go_goroutines"},
		},
		{
			name:       "filtered collectors",
			query:      "?collect[]=fan&collect[]=fan",
			wantStatus: http.StatusOK,
			wantBody:   []string{`rpi_scrape_collector_success{collector="fan"}`},
		},
		{
			name:            "unknown collector",
			query:           "?collect[]=%3Cb%3Edisk%3C%2Fb%3E",
			wantStatus:      http.StatusBadRequest,
			wantContentType: "text/plain; charset=utf-8",
			wantBody:        []string{`unknown collector "<b>disk</b>"`, "collect[] takes one of: cpu, fan, gpu, textfile"},
		},
	}
	h := newHandler(true)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics"+tt.query, http.NoBody))

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, w.Header().Get("Content-Type"))
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
			}
			for _, want := range tt.wantBody {
				assert.Contains(t, w.Body.String(), want)
			}
		})
	}

	// collect[]=fan, however often it's repeated, is one cache entry.
	assert.Len(t, h.filteredHandlers, 1)
}

func TestHandlerConcurrentScrapes(t *testing.T) {
	app := kingpin.New("rpi_exporter", "")
	collector.RegisterFlags(app)
	_, err := app.Parse(nil)
	require.NoError(t, err)

	// The race detector fails this test if the filtered handler cache isn't
	// guarded.
	h := newHandler(false)
	var wg sync.WaitGroup
	for _, filter := range []string{"cpu", "fan", "gpu", "textfile", "cpu", "fan"} {
		wg.Go(func() {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics?collect[]="+filter, http.NoBody))
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
	wg.Wait()
	assert.Len(t, h.filteredHandlers, 4)
}

func TestHealthCheckHandler(t *testing.T) {
	w := httptest.NewRecorder()
	HealthCheckHandler(w, httptest.NewRequest(http.MethodGet, "/health", http.NoBody))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	body, err := io.ReadAll(w.Body)
	require.NoError(t, err)
	assert.JSONEq(t, `{"alive": true}`, string(body))
}
