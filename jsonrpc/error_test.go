package jsonrpc

import (
	"encoding/json"
	"testing"
)

func TestNestedOdooErrorDetection(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		expectError  bool
	}{
		{
			name: "nested odoo error in result",
			responseBody: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": {
					"jsonrpc": "2.0",
					"id": null,
					"error": {
						"code": 200,
						"message": "Odoo Server Error",
						"data": {
							"name": "odoo.exceptions.UserError",
							"arguments": ["Non-stored field cannot be searched"],
							"context": {},
							"exception_type": "user"
						}
					}
				}
			}`,
			expectError: true,
		},
		{
			name: "normal successful response",
			responseBody: `{
				"jsonrpc": "2.0",
				"id": 1,
				"result": [{"id": 1, "name": "Test Product"}]
			}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a response struct and unmarshal the test data
			var rpcResp response
			err := json.Unmarshal([]byte(tt.responseBody), &rpcResp)
			if err != nil {
				t.Fatalf("Failed to unmarshal test response: %v", err)
			}

			// Test the nested error detection logic
			var result any
			if !tt.expectError {
				// Should not detect nested error in successful response
				var odooResp odooServerResponse
				err = json.Unmarshal(rpcResp.Result, &odooResp)
				if err == nil && odooResp.Error != nil {
					t.Errorf("Expected no nested error, but found one: %v", odooResp.Error)
				}
			} else {
				// Should detect nested error in error response
				var odooResp odooServerResponse
				err = json.Unmarshal(rpcResp.Result, &odooResp)
				if err != nil {
					t.Fatalf("Failed to detect nested error structure: %v", err)
				}
				if odooResp.Error == nil {
					t.Errorf("Expected nested error, but found none")
				} else {
					// Verify error creation works
					odooErr := newOdooError(odooResp.Error.Code, odooResp.Error.Message, odooResp.Error.Data)
					if odooErr == nil {
						t.Errorf("Expected error creation to return error, got nil")
					}
					t.Logf("Successfully detected nested error: %v", odooErr)
				}
			}

			// Test normal result processing
			if !tt.expectError {
				err = json.Unmarshal(rpcResp.Result, &result)
				if err != nil {
					t.Errorf("Failed to unmarshal normal result: %v", err)
				}
			}
		})
	}
}

func TestOdooError_Creation(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "user error",
			code:     200,
			message:  "Test error",
			data:     map[string]interface{}{"name": "odoo.exceptions.UserError"},
			expected: "odoo user error 200: Test error - map[name:odoo.exceptions.UserError]",
		},
		{
			name:     "validation error",
			code:     200,
			message:  "Validation failed",
			data:     map[string]interface{}{"name": "odoo.exceptions.ValidationError"},
			expected: "odoo validation error 200: Validation failed - map[name:odoo.exceptions.ValidationError]",
		},
		{
			name:     "access error",
			code:     200,
			message:  "Access denied",
			data:     map[string]interface{}{"name": "odoo.exceptions.AccessError"},
			expected: "odoo access error 200: Access denied - map[name:odoo.exceptions.AccessError]",
		},
		{
			name:     "generic error",
			code:     500,
			message:  "Server error",
			data:     map[string]interface{}{"name": "odoo.exceptions.Warning"},
			expected: "odoo error 500: Server error - map[name:odoo.exceptions.Warning]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newOdooError(tt.code, tt.message, tt.data)

			if err == nil {
				t.Errorf("expected error but got none")
				return
			}

			if err.Error() != tt.expected {
				t.Errorf("expected error message %q, got %q", tt.expected, err.Error())
			}
		})
	}
}
