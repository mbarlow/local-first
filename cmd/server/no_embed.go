//go:build !embed

package main

import "embed"

// These variables exist but are empty when not building with embed tags
var webFiles embed.FS
var hasEmbedded = false