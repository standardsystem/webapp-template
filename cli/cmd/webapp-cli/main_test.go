package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRun_version(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, []string{"version"})
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if stdout.String() != Version+"\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestRun_health_ok(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, []string{"health", srv.URL})
	if code != 0 {
		t.Fatalf("want exit 0, got %d", code)
	}
	if stdout.String() != "ok\n" {
		t.Fatalf("stdout: %q", stdout.String())
	}
}

func TestRun_health_badStatus(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, []string{"health", srv.URL})
	if code != 1 {
		t.Fatalf("want exit 1, got %d", code)
	}
}

func TestRun_noArgs(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, []string{})
	if code != 2 {
		t.Fatalf("want exit 2, got %d", code)
	}
}

func TestRun_unknownCommand(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	code := run(&stdout, &stderr, []string{"nope"})
	if code != 2 {
		t.Fatalf("want exit 2, got %d", code)
	}
}

func TestCheckHealth_invalidURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "empty", raw: "", wantErr: true},
		{name: "no scheme", raw: "localhost:8080/health", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkHealth(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkHealth(%q) err=%v wantErr=%v", tt.raw, err, tt.wantErr)
			}
		})
	}
}
