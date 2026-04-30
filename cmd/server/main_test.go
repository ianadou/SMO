package main

import (
	"errors"
	"slices"
	"testing"
)

func TestParsePort_ReturnsDefault_WhenInputIsEmpty(t *testing.T) {
	t.Parallel()

	got, err := parsePort("")
	if err != nil {
		t.Fatalf("expected no error on empty input, got %v", err)
	}
	if got != defaultPort {
		t.Errorf("expected default %q, got %q", defaultPort, got)
	}
}

func TestParsePort_ReturnsValue_WhenInputIsValid(t *testing.T) {
	t.Parallel()

	got, err := parsePort("9090")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "9090" {
		t.Errorf("expected '9090', got %q", got)
	}
}

func TestParsePort_ReturnsErrInvalidPort_OnNonNumeric(t *testing.T) {
	t.Parallel()

	_, err := parsePort("abc")
	if !errors.Is(err, errInvalidPort) {
		t.Errorf("expected errInvalidPort, got %v", err)
	}
}

func TestParsePort_ReturnsErrInvalidPort_OnOutOfRange(t *testing.T) {
	t.Parallel()

	cases := []string{"0", "-1", "65536", "70000"}
	for _, raw := range cases {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			_, err := parsePort(raw)
			if !errors.Is(err, errInvalidPort) {
				t.Errorf("parsePort(%q): expected errInvalidPort, got %v", raw, err)
			}
		})
	}
}

func TestParsePort_AcceptsBoundaries(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"1", "65535"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			got, err := parsePort(raw)
			if err != nil {
				t.Errorf("parsePort(%q): unexpected error %v", raw, err)
			}
			if got != raw {
				t.Errorf("parsePort(%q) = %q, want %q", raw, got, raw)
			}
		})
	}
}

func TestParseTrustedProxies_ReturnsDefaultRanges_WhenInputIsEmpty(t *testing.T) {
	t.Parallel()

	got := parseTrustedProxies("")
	want := []string{
		"127.0.0.1", "::1",
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	}
	if !slices.Equal(got, want) {
		t.Errorf("default proxies mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestParseTrustedProxies_SplitsCommaSeparatedCIDRs(t *testing.T) {
	t.Parallel()

	got := parseTrustedProxies("10.0.0.0/24,192.168.1.1")
	want := []string{"10.0.0.0/24", "192.168.1.1"}
	if !slices.Equal(got, want) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func TestParseTrustedProxies_TrimsWhitespaceAroundEntries(t *testing.T) {
	t.Parallel()

	got := parseTrustedProxies("  10.0.0.0/24 , 192.168.1.1  ")
	want := []string{"10.0.0.0/24", "192.168.1.1"}
	if !slices.Equal(got, want) {
		t.Errorf("expected trimmed entries, got=%v want=%v", got, want)
	}
}

func TestParseTrustedProxies_DropsEmptyEntries(t *testing.T) {
	t.Parallel()

	got := parseTrustedProxies("10.0.0.0/24,,192.168.1.1, ")
	want := []string{"10.0.0.0/24", "192.168.1.1"}
	if !slices.Equal(got, want) {
		t.Errorf("expected empty entries dropped, got=%v want=%v", got, want)
	}
}

func TestParseAllowedOrigins_ReturnsLocalhostDevPair_WhenInputIsEmpty(t *testing.T) {
	t.Parallel()

	got := parseAllowedOrigins("")
	want := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3001",
	}
	if !slices.Equal(got, want) {
		t.Errorf("default origins mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestParseAllowedOrigins_SplitsCommaSeparatedOrigins(t *testing.T) {
	t.Parallel()

	got := parseAllowedOrigins("https://sportpotes.fr,https://staging.sportpotes.fr")
	want := []string{"https://sportpotes.fr", "https://staging.sportpotes.fr"}
	if !slices.Equal(got, want) {
		t.Errorf("got=%v want=%v", got, want)
	}
}

func TestParseAllowedOrigins_TrimsWhitespaceAroundEntries(t *testing.T) {
	t.Parallel()

	got := parseAllowedOrigins("  https://sportpotes.fr , https://staging.sportpotes.fr  ")
	want := []string{"https://sportpotes.fr", "https://staging.sportpotes.fr"}
	if !slices.Equal(got, want) {
		t.Errorf("expected trimmed entries, got=%v want=%v", got, want)
	}
}

func TestParseAllowedOrigins_DropsEmptyEntries(t *testing.T) {
	t.Parallel()

	got := parseAllowedOrigins("https://sportpotes.fr,,https://staging.sportpotes.fr, ")
	want := []string{"https://sportpotes.fr", "https://staging.sportpotes.fr"}
	if !slices.Equal(got, want) {
		t.Errorf("expected empty entries dropped, got=%v want=%v", got, want)
	}
}
