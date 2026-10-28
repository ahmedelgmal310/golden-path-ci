package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, path string) (int, map[string]string) {
	t.Helper()
	rec := httptest.NewRecorder()
	newMux("1.2.3").ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("%s: invalid JSON %q: %v", path, rec.Body.String(), err)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("%s: content-type %q", path, ct)
	}
	return rec.Code, body
}

func TestHealthz(t *testing.T) {
	code, body := get(t, "/healthz")
	if code != http.StatusOK || body["status"] != "ok" {
		t.Fatalf("got %d %v", code, body)
	}
}

func TestHello(t *testing.T) {
	tests := []struct {
		name, query, want string
		code              int
	}{
		{"default", "/hello", "hello, world", http.StatusOK},
		{"named", "/hello?name=Ada", "hello, Ada", http.StatusOK},
		{"trimmed", "/hello?name=%20Ada%20", "hello, Ada", http.StatusOK},
		{"too long", "/hello?name=" + strings.Repeat("a", 65), "", http.StatusBadRequest},
		{"invalid utf8", "/hello?name=%ff", "", http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, body := get(t, tc.query)
			if code != tc.code {
				t.Fatalf("code %d, want %d (%v)", code, tc.code, body)
			}
			if tc.code == http.StatusOK {
				if body["message"] != tc.want || body["version"] != "1.2.3" {
					t.Fatalf("got %v", body)
				}
			} else if body["error"] == "" {
				t.Fatalf("expected an error message, got %v", body)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	newMux("x").ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/hello", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /hello returned %d", rec.Code)
	}
}
