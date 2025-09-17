package odoorpc

import (
	"context"
	"encoding/json"
)

type ServerVersion struct {
	ServerVersion     string      `json:"server_version"`
	ServerVersionInfo VersionInfo `json:"server_version_info"`
	ServerSerie       string      `json:"server_serie"`
	ProtocolVersion   int         `json:"protocol_version"`
}

// VersionInfo describes the structured data returned by Odoo's version endpoint.
type VersionInfo struct {
	Major        int
	Minor        int
	Patch        int
	ReleaseLevel string
	Serial       int
	FullVersion  string
}

// UnmarshalJSON maps the array returned by Odoo into the VersionInfo struct.
func (v *VersionInfo) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) == 0 {
		*v = VersionInfo{}
		return nil
	}

	var info VersionInfo
	for i, item := range raw {
		switch i {
		case 0:
			if err := json.Unmarshal(item, &info.Major); err != nil {
				return err
			}
		case 1:
			if err := json.Unmarshal(item, &info.Minor); err != nil {
				return err
			}
		case 2:
			if err := json.Unmarshal(item, &info.Patch); err != nil {
				return err
			}
		case 3:
			if err := json.Unmarshal(item, &info.ReleaseLevel); err != nil {
				return err
			}
		case 4:
			if err := json.Unmarshal(item, &info.Serial); err != nil {
				return err
			}
		case 5:
			if err := json.Unmarshal(item, &info.FullVersion); err != nil {
				return err
			}
		}
	}

	*v = info
	return nil
}

// Client defines the operations required to interact with Odoo.
type Client interface {
	// Version get metadata version of the server
	Version(ctx context.Context) (ServerVersion, error)
	// Authenticate logs in the user and returns its uid.
	Authenticate(ctx context.Context, username, password, db string) (int64, error)
	// SearchRead queries a model and return the fields.
	SearchRead(ctx context.Context, model string, domain Domain, opts Options) ([]map[string]any, error)
	// Search queries a model
	Search(ctx context.Context, model string, domain Domain, opts Options) ([]int64, error)
	// Create adds a new record to the given model and returns its ID.
	Create(ctx context.Context, model string, values map[string]any) (int64, error)
	// Update modifies fields for the specified records of a model.
	Update(ctx context.Context, model string, ids []int64, values map[string]any) (bool, error)
	// Unlink removes records from a model.
	Unlink(ctx context.Context, model string, ids []int64) (bool, error)
	// FieldsGet retrieves metadata for fields of a model.
	FieldsGet(ctx context.Context, model string, fields []string, opts Options) (map[string]any, error)
	// CallMethod model method
	CallMethod(ctx context.Context, model, method string, vars []any, opts Options) ([]any, error)
	// Read method
	Read(ctx context.Context, model string, ids []int64, opts Options) ([]map[string]any, error)
}
