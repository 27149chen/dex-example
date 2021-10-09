package main

import (
	"fmt"
	"os"

	"github.com/27149chen/dex-example/pkg/authn"
)

func main() {
	if err := authn.RootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
}
