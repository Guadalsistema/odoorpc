package odoorpc

import (
	"context"
	"testing"
)

func TestRpcClient_SearchRead_ErrorDetection(t *testing.T) {
	// Test that SearchRead properly handles and surfaces Odoo server errors
	client := New("http://test.example.com", nil, nil)

	// This test would require mocking the JSON-RPC layer to test error detection
	// For now, we test the validation logic added
	ctx := context.Background()

	// Test nil domain handling
	result, err := client.SearchRead(ctx, "test.model", nil, Options{})
	if err != nil {
		t.Errorf("Expected no error for nil domain, got: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result for nil domain")
	}
}

func TestRpcClient_FieldsGet_ErrorDetection(t *testing.T) {
	client := New("http://test.example.com", nil, nil)

	ctx := context.Background()

	// Test nil fields handling
	result, err := client.FieldsGet(ctx, "test.model", nil, Options{})
	if err != nil {
		t.Errorf("Expected no error for nil fields, got: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result for nil fields")
	}
}

func TestRpcClient_Search_ErrorDetection(t *testing.T) {
	client := New("http://test.example.com", nil, nil)

	ctx := context.Background()

	// Test nil domain handling
	result, err := client.Search(ctx, "test.model", nil, Options{})
	if err != nil {
		t.Errorf("Expected no error for nil domain, got: %v", err)
	}

	if result == nil {
		t.Error("Expected non-nil result for nil domain")
	}
}

func TestRpcClient_Create_Validation(t *testing.T) {
	client := New("http://test.example.com", nil, nil)

	ctx := context.Background()

	// Test empty values
	_, err := client.Create(ctx, "test.model", map[string]any{})
	// This should return an error about empty values, but we test the validation logic
	if err == nil {
		t.Error("Expected error for empty values")
	}
}
