package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Guadalsistema/odoorpc"
)

func main() {
	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.Options{Level: slog.LevelInfo}))
	client := odoorpc.New("https://odoo.example.com", nil, logger)

	ctx := context.Background()

	// Test the problematic scenario from the original issue
	domain := odoorpc.NewDomain().
		Equals("sale_ok", true).
		NotEquals("categ_id", false).
		NotEquals("image_1920", false) // This is the problematic filter

	opts := odoorpc.Options{
		Fields: []string{"name", "image_1920"},
		Limit:  10,
	}

	fmt.Println("Testing SearchRead with problematic domain...")
	products, err := client.SearchRead(ctx, "product.product", domain, opts)
	if err != nil {
		fmt.Printf("✅ SUCCESS: Error properly detected and surfaced!\n")
		fmt.Printf("Error: %v\n", err)

		// Show error type information
		if odoorpc.IsOdooError(err) {
			fmt.Printf("Error type: Odoo error detected\n")
			if odooErr, ok := odoorpc.GetOdooError(err); ok {
				fmt.Printf("Error code: %d\n", odooErr.Code)
				fmt.Printf("Error message: %s\n", odooErr.Message)
				if arguments, hasArgs := odooErr.Data["arguments"]; hasArgs {
					fmt.Printf("Error arguments: %v\n", arguments)
				}
			}
		}
		os.Exit(1)
	}

	fmt.Printf("❌ UNEXPECTED: No error returned (silent failure!)\n")
	fmt.Printf("Products returned: %d\n", len(products))
}
