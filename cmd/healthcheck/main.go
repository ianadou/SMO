// Command healthcheck probes the SMO server's /health/ready endpoint
// and exits 0 on HTTP 200, 1 otherwise. Required because the distroless
// runtime image has no shell — Dockerfile HEALTHCHECK cannot rely on
// curl/wget. /health/ready is the right target (not /health/live)
// because we want Docker to mark the container unhealthy when the
// database connection is lost, not just when the binary crashes.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultPort   = "8081"
	probeTimeout  = 2 * time.Second
	probePathName = "/health/ready"
)

func main() {
	os.Exit(run(os.Stderr, os.Getenv))
}

// run is the testable body of main. It returns the exit code instead
// of calling os.Exit directly so unit tests can assert on the result
// without terminating the test binary.
func run(stderr io.Writer, getenv func(string) string) int {
	port := resolvePort(getenv)
	url := fmt.Sprintf("http://127.0.0.1:%s%s", port, probePathName)

	client := &http.Client{Timeout: probeTimeout}
	resp, err := client.Get(url) //nolint:gosec // healthcheck always probes 127.0.0.1
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "healthcheck:", err)
		return 1
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = fmt.Fprintf(stderr, "healthcheck: expected 200, got %d\n", resp.StatusCode)
		return 1
	}
	return 0
}

// resolvePort returns the port the probe should hit: PORT from the
// environment, or defaultPort when unset. Extracted as a separate
// function so the fallback can be tested without depending on what
// happens to be listening on the default port.
func resolvePort(getenv func(string) string) string {
	if p := getenv("PORT"); p != "" {
		return p
	}
	return defaultPort
}
