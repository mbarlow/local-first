//go:build embed

package main

import "embed"

//go:embed web/*
var webFiles embed.FS
var hasEmbedded = true