package odoorpc

import "context"

type ServerVersion struct {
	ServerVersion     string `json:"server_version"`
	ServerVersionInfo []any  `json:"server_version_info"`
	ServerSerie       string `json:"server_serie"`
	ProtocolVersion   int    `json:"protocol_version"`
}

// Client defines the operations required to interact with Odoo.
type Client interface {
	// Version get metadata version of the server
	Version(ctx context.Context) (ServerVersion, error)
	// Authenticate logs in the user and returns its uid.
	Authenticate(ctx context.Context, username, password string) (int64, error)
	// SearchRead queries a model and return the fields.
	SearchRead(ctx context.Context, model string, domain Domain, opts Options) ([]map[string]any, error)
	// Search queries a model
	Search(ctx context.Context, model string, domain Domain, opts any) ([]int64, error)
	// Create adds a new record to the given model and returns its ID.
	Create(ctx context.Context, model string, values map[string]any) (int64, error)
	// Update modifies fields for the specified records of a model.
	Update(ctx context.Context, model string, ids []int64, values map[string]any) (bool, error)
	// Unlink removes records from a model.
	Unlink(ctx context.Context, model string, ids []int64) (bool, error)
	// FieldsGet retrieves metadata for fields of a model.
	FieldsGet(ctx context.Context, model string, fields []string) (map[string]any, error)
	// CallMethod model method
	CallMethod(ctx context.Context, model, method string, vars []any, kwargs map[string]any) ([]any, error)
	// Read method
	Read(ctx context.Context, model string, ids []int64, fields []string) ([]map[string]any, error)
}
