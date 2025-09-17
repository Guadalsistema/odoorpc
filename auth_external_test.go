//go:build odoo_external

package odoorpc_test

import (
	"context"
	"testing"

	"github.com/Guadalsistema/odoorpc"
)

// TestAuthenticateAgainstOdoo validates that admin authentication works against a local Odoo instance.
func TestAuthenticateAgainstOdoo(t *testing.T) {
	c := odoorpc.New("http://127.0.0.1:8069", nil)
	ctx := context.Background()

	resp, err := c.Version(ctx)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if resp.ServerVersionInfo.Major < 16 {
		t.Fatalf("Unexpected version answer %v", resp.ServerVersion)
	}

	uid, err := c.Authenticate(ctx, "admin", "admin", "odoo")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if uid <= 0 {
		t.Fatalf("unexpected uid %d", uid)
	}

	opts := odoorpc.Options{Fields: []string{"name"}}
	users, err := c.SearchRead(ctx, "res.users", nil, opts)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(users) == 0 {
		t.Fatalf("SearchRead no users returned")
	}

	uids, err := c.Search(ctx, "res.users", nil, odoorpc.Options{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(uids) == 0 {
		t.Fatalf("Search no users returned")
	}
}
