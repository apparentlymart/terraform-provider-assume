// terraform-provider-assume is a small utility provider that provides
// functions for telling Terraform to make various assumptions about unknown
// values.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apparentlymart/terraform-provider-assume/internal/assume"
)

func main() {
	provider := assume.NewProvider()
	err := provider.Serve(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start provider: %s", err)
		os.Exit(1)
	}
}
