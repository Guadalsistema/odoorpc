# odoorpc

Go client for Odoo JSON-RPC API with **enhanced error detection and handling**.

## 🚨 Breaking Changes in v2.0

This version introduces **breaking changes** to fix silent Odoo error failures:

- ❌ **Backward Compatibility**: Intentionally broken to surface silent errors
- ✅ **Error Detection**: Odoo server errors now properly surface as Go errors
- ✅ **New Error Types**: Specific error types for different Odoo exceptions

## Installation

```bash
go get github.com/Guadalsistema/odoorpc
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "github.com/Guadalsistema/odoorpc"
)

func main() {
    // Create client with optional logger
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.Options{Level: slog.LevelInfo}))
    client := odoorpc.New("https://odoo.example.com", nil, logger)

    ctx := context.Background()
    
    // Authenticate
    uid, err := client.Authenticate(ctx, "username", "password", "database")
    if err != nil {
        if odoorpc.IsOdooError(err) {
            // Handle Odoo-specific errors
            fmt.Printf("Odoo error: %v", err)
        } else {
            // Handle network/protocol errors
            fmt.Printf("Network error: %v", err)
        }
        return
    }

    // Search records
    products, err := client.SearchRead(ctx, "product.product", 
        odoorpc.NewDomain().Equals("sale_ok", true), 
        odoorpc.Options{Fields: []string{"name", "price"}})
    if err != nil {
        fmt.Printf("Search failed: %v", err)
        return
    }

    fmt.Printf("Found %d products", len(products))
}
```

## Key Features

### ✅ **Enhanced Error Detection**

Detects and properly surfaces **embedded Odoo server errors** that were previously silent:

```go
// This now properly surfaces "Non-stored field cannot be searched" errors
products, err := client.SearchRead(ctx, "product.product", 
    odoorpc.NewDomain().NotEquals("image_1920", false), opts)
if err != nil {
    // err is now: "odoo user error 200: Odoo Server Error - map[arguments:[Non-stored field cannot be searched]...]"
}
```

### ✅ **Specific Error Types**

- `OdooValidationError`: Data validation errors
- `OdooAccessError`: Permission/access errors  
- `OdooUserError`: User action errors
- `OdooError`: Generic Odoo errors

### ✅ **Response Validation**

All RPC methods include:
- Structure integrity checks
- Nil response detection
- Malformed data validation

### ✅ **Debug Logging**

Structured debugging with request/response details:
```bash
export LOG_LEVEL=DEBUG
./your-app
```

## API Reference

### Client Methods

| Method | Description | Enhanced |
|--------|-------------|-----------|
| `Version()` | Get server version | ✅ Error detection |
| `Authenticate()` | User authentication | ✅ Enhanced validation |
| `SearchRead()` | Query with domain | ✅ Embedded error detection |
| `Search()` | Get record IDs | ✅ Response validation |
| `FieldsGet()` | Field metadata | ✅ Structure validation |
| `Create()` | Create record | ✅ Result validation |
| `Update()` | Update records | ✅ Debug logging |
| `Unlink()` | Delete records | ✅ Error handling |
| `Read()` | Read by IDs | ✅ Validation |
| `CallMethod()` | Custom methods | ✅ Error mapping |

### Domain Building

```go
// Build search domains fluently
domain := odoorpc.NewDomain().
    Equals("active", true).
    NotEquals("categ_id", false).
    GreaterThan("create_date", "2023-01-01")
```

### Error Handling

```go
if err != nil {
    switch e := err.(type) {
    case *odoorpc.OdooValidationError:
        // Handle validation errors
    case *odoorpc.OdooAccessError:
        // Handle permission errors
    case *odoorpc.OdooUserError:
        // Handle user errors
    default:
        if odoorpc.IsOdooError(err) {
            // Generic Odoo error
            odooErr, _ := odoorpc.GetOdooError(err)
            fmt.Printf("Odoo error: %s", odooErr.Message)
        }
    }
}
```

## Migration from v1.x

### Breaking Changes

1. **Error Behavior**: Silent failures now surface as errors
2. **Error Types**: New specific error types introduced
3. **Response Validation**: Additional validation for all methods

### Update Your Code

```go
// OLD: Might miss Odoo server errors
if err != nil {
    log.Fatal(err)
}

// NEW: Comprehensive error handling
if err != nil {
    if odoorpc.IsOdooError(err) {
        // Handle Odoo-specific errors
        odooErr, _ := odoorpc.GetOdooError(err)
        log.Printf("Odoo error: %s", odooErr.Message)
    } else {
        // Handle other errors
        log.Printf("Network error: %v", err)
    }
}
```

## Examples

### E-commerce Integration

```go
// Sync products from Odoo
func syncProducts(ctx context.Context, client *odoorpc.RpcClient) error {
    // Get active products
    products, err := client.SearchRead(ctx, "product.product", 
        odoorpc.NewDomain().
            Equals("sale_ok", true).
            NotEquals("image_1920", false),  // Will now properly surface error
        odoorpc.Options{
            Fields: []string{"name", "description", "price", "image_1920"},
            Limit: 100,
        })
    
    if err != nil {
        return fmt.Errorf("failed to fetch products: %w", err)
    }

    // Process products...
    return nil
}
```

### Inventory Management

```go
// Create new product
func createProduct(ctx context.Context, client *odoorpc.RpcClient, product map[string]any) (int64, error) {
    id, err := client.Create(ctx, "product.product", product)
    if err != nil {
        if odoorpc.IsOdooError(err) {
            // Handle validation errors like missing required fields
            if odooErr, ok := odoorpc.GetOdooError(err); ok {
                return 0, fmt.Errorf("validation failed: %s", odooErr.Message)
            }
        }
        return 0, fmt.Errorf("create failed: %w", err)
    }
    
    return id, nil
}
```

## Development

### Running Tests

```bash
go test ./jsonrpc -v    # JSON-RPC layer tests
go test ./... -v      # All tests
```

### Building

```bash
go build .
go build -o odoorpc-client ./examples/
```

## Issues This Fixes

### 🐛 **Silent Odoo Error Problem**

**Before**: Odoo server errors were silent:
```bash
# Query with invalid field
client.SearchRead(ctx, "product.product", domain.NotEquals("image_1920", false), opts)
# Result: No error, empty slice (silent failure)
# Server logs: "Non-stored field cannot be searched"
```

**After**: Proper error surface:
```bash
client.SearchRead(ctx, "product.product", domain.NotEquals("image_1920", false), opts)
# Result: error = "odoo user error 200: Odoo Server Error - map[arguments:[Non-stored field cannot be searched]...]"
```

This provides:
- ✅ **Clear Error Messages**: No more silent failures
- ✅ **Debug Information**: Full error context with data
- ✅ **Error Classification**: Specific error types for handling
- ✅ **Production Ready**: Structured logging integration

## License

[View License](LICENSE)

## Contributing

1. Fork the repository
2. Create feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request

## Related Projects

- [odoo-client](https://github.com/Guadalsistema/connector-proyect) - Full connector framework using this library
- [connector-proyect](https://github.com/Guadalsistema/connector-proyect) - Odoo integration examples