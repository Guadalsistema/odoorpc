package jsonrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync/atomic"
)

// Client is a minimal JSON-RPC 2.0 client.
type NetClient struct {
	endpoint   string
	httpClient *http.Client
	nextID     uint64
}

// New creates a new JSON-RPC client for the given endpoint.
func New(endpoint string, httpClient *http.Client) *NetClient {
	// Ensure endpoint ends with /jsonrpc
	if !strings.HasSuffix(endpoint, "/jsonrpc") {
		endpoint = strings.TrimRight(endpoint, "/") + "/jsonrpc"
	}

	if httpClient == nil {
		jar, _ := cookiejar.New(nil)
		httpClient = &http.Client{Jar: jar}
	} else if httpClient.Jar == nil {
		jar, _ := cookiejar.New(nil)
		httpClient.Jar = jar
	}

	return &NetClient{endpoint: endpoint, httpClient: httpClient}
}

type request struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      uint64 `json:"id"`
}

type rpcError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
	ID      uint64          `json:"id"`
}

// OdooError represents an error response from Odoo server (embedded in result)
type odooError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

// OdooServerResponse represents a response from Odoo server that may contain nested errors
type odooServerResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   *odooError  `json:"error"`
	Result  interface{} `json:"result"`
}

func (e *odooError) Error() string {
	return fmt.Sprintf("odoo error %d: %s", e.Code, e.Message)
}

func newOdooError(code int, message string, data map[string]interface{}) error {
	if data == nil {
		data = make(map[string]interface{})
	}

	// Extract exception type from data if available
	exceptionType, hasType := data["name"].(string)
	if hasType {
		switch exceptionType {
		case "odoo.exceptions.ValidationError":
			return fmt.Errorf("odoo validation error %d: %s - %v", code, message, data)
		case "odoo.exceptions.AccessError":
			return fmt.Errorf("odoo access error %d: %s - %v", code, message, data)
		case "odoo.exceptions.UserError":
			return fmt.Errorf("odoo user error %d: %s - %v", code, message, data)
		default:
			return fmt.Errorf("odoo error %d: %s - %v", code, message, data)
		}
	}

	return fmt.Errorf("odoo error %d: %s - %v", code, message, data)
}

// Call performs a JSON-RPC request and decodes the result into result.
func (c *NetClient) Call(ctx context.Context, method string, params any, result any) error {
	id := atomic.AddUint64(&c.nextID, 1)
	reqBody, err := json.Marshal(request{JSONRPC: "2.0", Method: method, Params: params, ID: id})
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %q: %w", method, err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request error to %s: %w", c.endpoint, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jsonrpc error response status: %d body: %s", resp.StatusCode, string(body))
	}

	var rpcResp response
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		wrapped := fmt.Errorf("jsonrpc decode failed: %w", err)
		return fmt.Errorf("jsonrpc decode failed: %v; body: %s", wrapped, snippet)
	}

	// Check for JSON-RPC errors
	if rpcResp.Error != nil {
		data, err := json.Marshal(rpcResp.Error.Data)
		if err != nil {
			return fmt.Errorf("jsonrpc error %d: %s (failed to marshal error data: %v)", rpcResp.Error.Code, rpcResp.Error.Message, err)
		}
		return fmt.Errorf("jsonrpc error %d: %s - %s", rpcResp.Error.Code, rpcResp.Error.Message, string(data))
	}

	// Check for nested Odoo errors in result payload
	if result != nil {
		// First, try to detect if the result contains an embedded Odoo error response
		var odooResp odooServerResponse
		if err := json.Unmarshal(rpcResp.Result, &odooResp); err == nil && odooResp.Error != nil {
			// Found nested Odoo error - convert to proper Go error
			return newOdooError(odooResp.Error.Code, odooResp.Error.Message, odooResp.Error.Data)
		}

		// Normal result processing
		if err := json.Unmarshal(rpcResp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}
	return nil
}
