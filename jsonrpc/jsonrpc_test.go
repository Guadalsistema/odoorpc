package jsonrpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testResult struct {
	Value string `json:"value"`
}

func TestClientCall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]any
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      req["id"],
			"result":  testResult{Value: "ok"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := New(srv.URL, srv.Client())
	var res testResult
	if err := c.Call(context.Background(), "test", nil, &res); err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if res.Value != "ok" {
		t.Fatalf("unexpected result: %v", res.Value)
	}
}

func TestNewEnsuresSuffixAndJar(t *testing.T) {
	httpClient := &http.Client{}
	c := New("http://example.com", httpClient)
	if !strings.HasSuffix(c.endpoint, "/jsonrpc") {
		t.Fatalf("expected endpoint to end with /jsonrpc got %s", c.endpoint)
	}
	if httpClient.Jar == nil {
		t.Fatalf("expected jar to be initialized")
	}
}

func TestCallHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("bad"))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	if err := c.Call(context.Background(), "method", nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCallRPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"error": map[string]any{
				"code":    -1,
				"message": "boom",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	if err := c.Call(context.Background(), "m", nil, nil); err == nil {
		t.Fatalf("expected rpc error")
	}
}

func TestCallInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())
	if err := c.Call(context.Background(), "m", nil, nil); err == nil {
		t.Fatalf("expected decode error")
	}
}
