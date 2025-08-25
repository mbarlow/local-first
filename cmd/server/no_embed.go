//go:build !embed

package main

import "embed"

var webFiles embed.FS
var hasEmbedded = false