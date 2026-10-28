package main

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRunShutsDownOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, "127.0.0.1:0", slog.New(slog.NewTextHandler(io.Discard, nil))) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run did not return after cancel")
	}
}

func TestRunInvalidAddress(t *testing.T) {
	if err := run(context.Background(), "127.0.0.1:-1", slog.New(slog.NewTextHandler(io.Discard, nil))); err == nil {
		t.Fatal("expected a listen error")
	}
}
