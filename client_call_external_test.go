//go:build odoo_external

package odoorpc_test

import (
	"context"
	"testing"

	"github.com/Guadalsistema/odoorpc"
)

// TestRPCClientCallAgainstOdoo verifies that RPCClient.Call works with a real Odoo server.
func TestRPCClientCallAgainstOdoo(t *testing.T) {
	ctx := context.Background()
	resp, err := c.Version(ctx)
	if err != nil {
		t.Logf("logs:\n%s", logBuf.String())
		t.Fatalf("Version: %v", err)
	}
	if _, err := c.Authenticate(ctx, user, pass); err != nil {
		t.Logf("logs:\n%s", logBuf.String())
		t.Fatalf("Authenticate: %v", err)
	}

	domain := odoorpc.NewDomain()
	res, err := c.SearchRead(ctx, "res.users", domain, odoorpc.Options.SetFields("name"))
	if err != nil {
		t.Logf("logs:\n%s", logBuf.String())
		t.Fatalf("Call: %v", err)
	}
	if len(res) == 0 {
		t.Logf("logs:\n%s", logBuf.String())
		t.Fatalf("Call returned no users")
	}
	for _, m := range res {
		if _, ok := m["name"]; !ok {
			t.Fatalf("missing name field: %v", m)
		}
	}
}
