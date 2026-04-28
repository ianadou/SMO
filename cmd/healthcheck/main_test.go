package main

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// startProbedServer returns an httptest.Server bound to 127.0.0.1 with
// the given handler, plus the port it ended up on. The healthcheck
// binary always probes 127.0.0.1:$PORT/health/ready, so the test must
// pin the listener to that interface.
func startProbedServer(t *testing.T, handler http.HandlerFunc) (string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := httptest.NewUnstartedServer(handler)
	srv.Listener = listener
	srv.Start()

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		srv.Close()
		t.Fatalf("parse server URL: %v", err)
	}
	return parsed.Port(), srv.Close
}

// envOnly returns a getenv-shaped function that resolves only the
// PORT variable. Other variables resolve to the empty string. The
// signature mirrors os.Getenv so tests can swap it in directly.
func envOnly(_, value string) func(string) string {
	return func(name string) string {
		if name == "PORT" {
			return value
		}
		return ""
	}
}

func TestRun_Returns0_WhenServerReturns200(t *testing.T) {
	t.Parallel()

	port, stop := startProbedServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != probePathName {
			t.Errorf("expected probe path %q, got %q", probePathName, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer stop()

	var stderr bytes.Buffer
	if got := run(&stderr, envOnly("PORT", port)); got != 0 {
		t.Errorf("expected exit 0, got %d (stderr=%q)", got, stderr.String())
	}
}

func TestRun_Returns1_WhenServerReturnsNon200(t *testing.T) {
	t.Parallel()

	port, stop := startProbedServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	defer stop()

	var stderr bytes.Buffer
	if got := run(&stderr, envOnly("PORT", port)); got != 1 {
		t.Errorf("expected exit 1, got %d", got)
	}
	if !strings.Contains(stderr.String(), "expected 200, got 503") {
		t.Errorf("expected stderr to mention status 503, got %q", stderr.String())
	}
}

func TestRun_Returns1_WhenServerIsUnreachable(t *testing.T) {
	t.Parallel()

	// Port 1 is reserved (tcpmux). Connecting fast-fails on every
	// platform we run on, so the test stays under the probe timeout.
	var stderr bytes.Buffer
	if got := run(&stderr, envOnly("PORT", "1")); got != 1 {
		t.Errorf("expected exit 1 on transport error, got %d", got)
	}
	if !strings.Contains(stderr.String(), "healthcheck:") {
		t.Errorf("expected stderr to start with 'healthcheck:', got %q", stderr.String())
	}
}

func TestResolvePort_ReturnsEnvValue_WhenSet(t *testing.T) {
	t.Parallel()

	got := resolvePort(envOnly("PORT", "9090"))
	if got != "9090" {
		t.Errorf("expected '9090', got %q", got)
	}
}

func TestResolvePort_ReturnsDefault_WhenEnvIsEmpty(t *testing.T) {
	t.Parallel()

	got := resolvePort(func(string) string { return "" })
	if got != defaultPort {
		t.Errorf("expected default %q, got %q", defaultPort, got)
	}
}
