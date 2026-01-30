package odoorpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Guadalsistema/odoorpc/jsonrpc"
)

// RpcClient implements Client interface using JSON-RPC API.
type RpcClient struct {
	logger   *slog.Logger
	rpc      *jsonrpc.NetClient
	db       string
	uid      int64
	password string
}

// New creates a new RPCClient using provided url and database name.
func New(url string, httpClient *http.Client, logger *slog.Logger) *RpcClient {
	return &RpcClient{rpc: jsonrpc.New(url, httpClient), logger: logger}
}

// Version get metadata call
func (c *RpcClient) Version(ctx context.Context) (ServerVersion, error) {
	params := map[string]any{
		"service": "common",
		"method":  "version",
		"args":    []any{},
	}
	var res ServerVersion
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return ServerVersion{}, fmt.Errorf("Version failed: %w", err)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Version completed", "server_version", res.ServerVersion)
	}

	return res, nil
}

// Authenticate logs in the user and returns its uid.
func (c *RpcClient) Authenticate(ctx context.Context, username, password, db string) (int64, error) {
	params := map[string]any{
		"service": "common",
		"method":  "login",
		"args":    []any{db, username, password},
	}

	var ret any
	if err := c.rpc.Call(ctx, "call", params, &ret); err != nil {
		return 0, fmt.Errorf("Authenticate failed: %w", err)
	}

	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Authentication response", "response", ret)
	}

	var uid int64
	switch v := ret.(type) {
	case int64:
		uid = v
	case float64:
		uid = int64(v)
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return 0, ErrInvalidAuthResponse
		}
		uid = parsed
	default:
		return 0, ErrInvalidAuthResponse
	}

	c.password = password
	c.uid = uid
	c.db = db

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Authentication successful", "uid", uid, "database", db)
	}

	return uid, nil
}

// SearchRead queries an Odoo model using `execute_kw` RPC call with `search_read` method.
//
// It allows retrieving records from a specific model that match a given search domain,
// returning only the requested fields.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control.
//   - model: The Odoo model name to query (e.g., "res.partner").
//   - domain: A search domain built via Domain helpers.
//     Example: NewDomain().Equals("is_company", true).Equals("country_id", 1)
//   - fields: A list of field names to include in result.
//
// Returns:
//   - A slice of maps, where each map represents a record with key-value pairs for the requested fields.
//   - An error if the RPC call fails, returns invalid data, or contains Odoo server errors.
//
// Example:
//
//	     partners, err := client.SearchRead(ctx, "res.partner",
//	         NewDomain().Equals("is_company", true),
//	         []string{"name", "country_id"})
//		if err != nil {
//		    log.Fatal(err)
//		}
//		for _, partner := range partners {
//		    fmt.Println("Partner:", partner["name"])
//		}
//
// Internals:
//   - Constructs an RPC request to `object` service with method `execute_kw`.
//   - Uses `search_read` to fetch both matching records and their field values in one step.
//   - Automatically serializes `domain` and `fields` to match Odoo's JSON-RPC expectations.
//   - Validates response structure and detects embedded Odoo server errors.
func (c *RpcClient) SearchRead(ctx context.Context, model string, domain Domain, opts Options) ([]map[string]any, error) {
	if domain == nil {
		domain = Domain{}
	}
	args := []any{domain}
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "search_read", args, opts.Kwargs()},
	}
	var res []map[string]any
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return nil, fmt.Errorf("SearchRead failed: %w", err)
	}

	// Validate response structure
	if res == nil {
		return nil, fmt.Errorf("SearchRead returned nil result for model %s", model)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "SearchRead completed",
			"model", model,
			"result_count", len(res),
			"domain", fmt.Sprintf("%v", []any(domain)))
	}

	return res, nil
}

// FieldsGet retrieves metadata for specified fields of a model.
//
// It wraps the `fields_get` RPC method, returning a map whose keys are the
// field names and values contain their definitions.
func (c *RpcClient) FieldsGet(ctx context.Context, model string, fields []string, opts Options) (map[string]any, error) {
	if fields == nil {
		fields = []string{}
	}
	args := []any{fields}
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "fields_get", args, opts.Kwargs()},
	}
	var res map[string]any
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return nil, fmt.Errorf("FieldsGet failed: %w", err)
	}

	// Validate response structure
	if res == nil {
		return nil, fmt.Errorf("FieldsGet returned nil result for model %s", model)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "FieldsGet completed",
			"model", model,
			"field_count", len(fields),
			"result_keys", len(res))
	}

	return res, nil
}

// Search queries a model using the `search` method via `execute_kw`.
func (c *RpcClient) Search(ctx context.Context, model string, domain Domain, opts Options) ([]int64, error) {
	if domain == nil {
		domain = Domain{}
	}
	args := []any{domain}
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "search", args, opts.Kwargs()},
	}
	var res []int64
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return nil, fmt.Errorf("Search failed: %w", err)
	}

	// Validate response structure
	if res == nil {
		return nil, fmt.Errorf("Search returned nil result for model %s", model)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Search completed",
			"model", model,
			"result_count", len(res),
			"domain", fmt.Sprintf("%v", []any(domain)))
	}

	return res, nil
}

// Create adds a new record to the given model and returns its ID.
func (c *RpcClient) Create(ctx context.Context, model string, values map[string]any) (int64, error) {
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "create", []any{values}},
	}
	var id int64
	if err := c.rpc.Call(ctx, "call", params, &id); err != nil {
		return 0, fmt.Errorf("Create failed: %w", err)
	}

	// Validate response structure
	if id == 0 {
		return 0, fmt.Errorf("Create returned zero ID for model %s", model)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Create completed",
			"model", model,
			"record_id", id,
			"field_count", len(values))
	}

	return id, nil
}

// Update modifies fields for specified records of a model.
func (c *RpcClient) Update(ctx context.Context, model string, ids []int64, values map[string]any) (bool, error) {
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "write", []any{ids, values}},
	}
	var res bool
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return false, fmt.Errorf("Update failed: %w", err)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Update completed",
			"model", model,
			"record_count", len(ids),
			"field_count", len(values),
			"result", res)
	}

	return res, nil
}

// Unlink removes records from a model.
func (c *RpcClient) Unlink(ctx context.Context, model string, ids []int64) (bool, error) {
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "unlink", []any{ids}},
	}
	var res bool
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return false, fmt.Errorf("Unlink failed: %w", err)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Unlink completed",
			"model", model,
			"record_count", len(ids),
			"result", res)
	}

	return res, nil
}

// CallMethod invokes an arbitrary method on the given model using JSON-RPC `call` service.
// vars contains positional arguments for the method, while kwargs holds keyword arguments.
//
// Some Odoo methods return scalar values (e.g. bool) instead of a list. To provide
// a consistent result type, non-slice responses are wrapped in a single-element
// slice before returning.
func (c *RpcClient) CallMethod(ctx context.Context, model, method string, vars []any, opts Options) ([]any, error) {
	if vars == nil {
		vars = []any{}
	}
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, method, vars, opts.Kwargs()},
	}
	var raw any
	if err := c.rpc.Call(ctx, "call", params, &raw); err != nil {
		return nil, fmt.Errorf("CallMethod failed: %w", err)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "CallMethod completed",
			"model", model,
			"method", method,
			"arg_count", len(vars),
			"result_type", fmt.Sprintf("%T", raw))
	}

	switch v := raw.(type) {
	case []any:
		return v, nil
	default:
		if v == nil {
			return nil, nil
		}
		return []any{v}, nil
	}
}

// Read fetches records by IDs from a model, optionally limited to specific fields.
// It returns a slice of maps, one per record (same shape as search_read).
func (c *RpcClient) Read(ctx context.Context, model string, ids []int64, opts Options) ([]map[string]any, error) {
	// If no IDs, nothing to read
	if len(ids) == 0 {
		return []map[string]any{}, nil
	}

	// Odoo expects a list of IDs as a positional arg: [ids]
	idArgs := make([]any, len(ids))
	for i, id := range ids {
		idArgs[i] = id
	}
	args := []any{idArgs}

	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "read", args, opts.Kwargs()},
	}

	var res []map[string]any
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return nil, fmt.Errorf("Read failed: %w", err)
	}

	// Validate response structure
	if res == nil {
		return nil, fmt.Errorf("Read returned nil result for model %s", model)
	}

	// Log debug info if logger is available
	if c.logger != nil && c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.DebugContext(ctx, "Read completed",
			"model", model,
			"id_count", len(ids),
			"result_count", len(res))
	}

	return res, nil
}

// Assertion
var _ Client = (*RpcClient)(nil)
