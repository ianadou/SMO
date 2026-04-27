// Command healthcheck probes the SMO server's /health endpoint and exits 0
// on HTTP 200, 1 otherwise. Required because the distroless runtime image
// has no shell — Dockerfile HEALTHCHECK cannot rely on curl/wget.
package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	url := fmt.Sprintf("http://127.0.0.1:%s/health", port)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url) //nolint:gosec // healthcheck always probes 127.0.0.1
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		os.Exit(1)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck: expected 200, got %d\n", resp.StatusCode)
		os.Exit(1)
	}
}
