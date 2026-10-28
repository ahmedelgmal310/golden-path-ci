package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"
)

const maxNameLen = 64

func newMux(version string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		name, err := greetingName(r.URL.Query().Get("name"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "hello, " + name, "version": version})
	})
	return mux
}

type validationError string

func (e validationError) Error() string { return string(e) }

// greetingName normalises the optional name parameter.
func greetingName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "world", nil
	}
	if !utf8.ValidString(name) {
		return "", validationError("name must be valid UTF-8")
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		return "", validationError("name is longer than 64 characters")
	}
	return name, nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
