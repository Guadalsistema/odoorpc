//go:build odoo_external

package odoorpc_test

import (
	"context"
	"testing"

	"github.com/Guadalsistema/odoorpc"
)

// TestRpcClientCallAgainstOdoo verifies that RPCClient.Call works with a real Odoo server.
func TestRpcClientCallAgainstOdoo(t *testing.T) {
	ctx := context.Background()
	c := odoorpc.New("https://127.0.0.1", nil)

	_, err := c.Version(ctx)
	if err != nil {
		t.Fatalf("Version: %v", err)
	}
	if _, err := c.Authenticate(ctx, "admin", "admin", "odoo"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	domain := odoorpc.NewDomain()
	opts := odoorpc.Options{Fields: []string{"name"}}
	res, err := c.SearchRead(ctx, "res.users", domain, opts)
	if err != nil {
		t.Fatalf("Call: %v", err)
	}
	if len(res) == 0 {
		t.Fatalf("Call returned no users")
	}
	for _, m := range res {
		if _, ok := m["name"]; !ok {
			t.Fatalf("missing name field: %v", m)
		}
	}
}
