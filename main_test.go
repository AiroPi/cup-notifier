package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apprise "github.com/unraid/apprise-go"
)

func TestCheckForUpdatesReturnsErrorOnInvalidJSON(t *testing.T) {
	cache = make(map[string]map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	t.Setenv("CUP_URL", server.URL)
	t.Setenv("INSECURE_SKIP_VERIFY", "false")

	notifier := apprise.New()
	if err := checkForUpdates(notifier); err == nil {
		t.Fatal("expected an error for invalid JSON response")
	}
}

func TestCheckForUpdatesReturnsErrorOnMalformedResponse(t *testing.T) {
	cache = make(map[string]map[string]any)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"images":[{"parts":"invalid"}]}`))
	}))
	defer server.Close()

	t.Setenv("CUP_URL", server.URL)
	t.Setenv("INSECURE_SKIP_VERIFY", "false")

	notifier := apprise.New()
	if err := checkForUpdates(notifier); err == nil {
		t.Fatal("expected an error for malformed CUP response")
	}
}
