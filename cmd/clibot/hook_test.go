package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestHookNotifier_Notify_Success tests successful hook notification
func TestHookNotifier_Notify_Success(t *testing.T) {
	// Create a test server that accepts POST requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/octet-stream", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := &HookNotifier{timeout: 5 * time.Second}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(context.Background(), server.URL, data)
	assert.NoError(t, err)
}

// TestHookNotifier_Notify_WrongStatusCode tests handling of non-200 status codes
func TestHookNotifier_Notify_WrongStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := &HookNotifier{timeout: 5 * time.Second}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(context.Background(), server.URL, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status: 500")
}

// TestHookNotifier_Notify_Timeout tests timeout behavior
func TestHookNotifier_Notify_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := &HookNotifier{timeout: 100 * time.Millisecond}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(context.Background(), server.URL, data)
	assert.Error(t, err)
}

// TestHookNotifier_Notify_ContextCancellation tests context cancellation
func TestHookNotifier_Notify_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	notifier := &HookNotifier{timeout: 5 * time.Second}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(ctx, server.URL, data)
	assert.Error(t, err)
}

// TestHookNotifier_Notify_InvalidURL tests handling of invalid URLs
func TestHookNotifier_Notify_InvalidURL(t *testing.T) {
	notifier := &HookNotifier{timeout: 5 * time.Second}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(context.Background(), "://invalid-url", data)
	assert.Error(t, err)
}

// TestHookNotifier_Notify_ServerNotResponding tests handling of server not responding
func TestHookNotifier_Notify_ServerNotResponding(t *testing.T) {
	// Use a URL that's unlikely to have a server listening
	notifier := &HookNotifier{timeout: 100 * time.Millisecond}
	data := []byte(`{"test": "data"}`)

	err := notifier.Notify(context.Background(), "http://localhost:9999", data)
	assert.Error(t, err)
}
