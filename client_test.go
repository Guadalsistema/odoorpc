package odoorpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticateReturnsErrorWhenCallFails(t *testing.T) {
	t.Parallel()

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/jsonrpc" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"boom"}`))
	}))
	defer server.Close()

	client := New(server.URL, server.Client(), nil)

	uid, err := client.Authenticate(context.Background(), "user", "pass", "db")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if uid != 0 {
		t.Fatalf("expected uid to be zero on error, got %d", uid)
	}
	if calls != 1 {
		t.Fatalf("expected one call to jsonrpc endpoint, got %d", calls)
	}
	if client.uid != 0 || client.db != "" || client.password != "" {
		t.Fatalf("client state should not change on authentication failure")
	}
}
