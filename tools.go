//go:build tools
// +build tools

// Package tools is a place to put any tooling dependencies as imports.
// Go modules will be forced to download and install them.
package tools

import (
	// To generate mocks
	_ "github.com/golang/mock/mockgen"
)
