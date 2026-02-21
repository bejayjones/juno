// Package webui exposes the compiled SvelteKit build for go:embed.
// Run `npm run build` inside web/ before compiling the Go binary.
package webui

import "embed"

// FS holds the compiled SvelteKit output produced by `npm run build`.
// The build directory is at web/build/ relative to this file.
//
//go:embed all:build
var FS embed.FS
