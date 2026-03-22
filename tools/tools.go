//go:build tools
// +build tools

//go:generate bash -c "go build -ldflags \"-X 'github.com/vektra/mockery/v2/pkg/config.SemVer=v0.0.0-dev'\" -o ../bin/mockery github.com/vektra/mockery/v2"

// Package tools contains go:generate commands for all project tools with versions stored in local go.mod file
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
package tools

import (
	_ "github.com/sqlc-dev/sqlc/cmd/sqlc"
	_ "github.com/vektra/mockery/v2"
)
