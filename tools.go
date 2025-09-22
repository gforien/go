//go:build tools

package main

import (
    _ "github.com/golangci/golangci-lint/cmd/golangci-lint"
    _ "honnef.co/go/tools/cmd/staticcheck"
)
