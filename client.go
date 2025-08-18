package odoorpc

import (
	"context"
	"net/http"

	"github.com/Guadalsistema/odoorpc/jsonrpc"
)

// RpcClient implements the Client interface using the JSON-RPC API.
type RpcClient struct {
	rpc      *jsonrpc.NetClient
	db       string
	uid      int64
	password string
}

// New creates a new RPCClient using the provided url and database name.
func New(url, db string, httpClient *http.Client) *RpcClient {
	return &RpcClient{rpc: jsonrpc.New(url, httpClient), db: db}
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
		return ServerVersion{}, err
	}
	return res, nil
}

// Authenticate logs in the user and returns its uid.
func (c *RpcClient) Authenticate(ctx context.Context, username, password string) (int64, error) {
	params := map[string]any{
		"service": "common",
		"method":  "login",
		"args":    []any{c.db, username, password},
	}
	var uid int64
	if err := c.rpc.Call(ctx, "call", params, &uid); err != nil {
		return 0, err
	}
	c.password = password
	c.uid = uid
	return uid, nil
}

// SearchRead queries an Odoo model using the `execute_kw` RPC call with the `search_read` method.
//
// It allows retrieving records from a specific model that match a given search domain,
// returning only the requested fields.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout control.
//   - model: The Odoo model name to query (e.g., "res.partner").
//   - domain: A search domain built via Domain helpers.
//     Example: NewDomain().Equals("is_company", true).Equals("country_id", 1)
//   - fields: A list of field names to include in the result.
//
// Returns:
//   - A slice of maps, where each map represents a record with key-value pairs for the requested fields.
//   - An error if the RPC call fails or returns invalid data.
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
//   - Constructs an RPC request to the `object` service with method `execute_kw`.
//   - Uses `search_read` to fetch both matching records and their field values in one step.
//   - Automatically serializes the `domain` and `fields` to match Odoo's JSON-RPC expectations.
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
		return nil, err
	}
	return res, nil
}

// FieldsGet retrieves metadata for the specified fields of a model.
//
// It wraps the `fields_get` RPC method, returning a map whose keys are the
// field names and values contain their definitions.
func (c *RpcClient) FieldsGet(ctx context.Context, model string, fields []string) (map[string]any, error) {
	if fields == nil {
		fields = []string{}
	}
	args := []any{fields}
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "fields_get", args, map[string]any{}},
	}
	var res map[string]any
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return nil, err
	}
	return res, nil
}

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
		return nil, err
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
		return 0, err
	}
	return id, nil
}

// Update modifies fields for the specified records of a model.
func (c *RpcClient) Update(ctx context.Context, model string, ids []int64, values map[string]any) (bool, error) {
	params := map[string]any{
		"service": "object",
		"method":  "execute_kw",
		"args":    []any{c.db, c.uid, c.password, model, "write", []any{ids, values}},
	}
	var res bool
	if err := c.rpc.Call(ctx, "call", params, &res); err != nil {
		return false, err
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
		return false, err
	}
	return res, nil
}

// CallMethod invokes an arbitrary method on the given model using the JSON-RPC `call` service.
// vars contains the positional arguments for the method, while kwargs holds keyword arguments.
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
		return nil, err
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
		return nil, err
	}
	return res, nil
}

// Assertion
var _ Client = (*RpcClient)(nil)
