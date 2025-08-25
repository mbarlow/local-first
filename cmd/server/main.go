package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	var (
		port      = flag.String("port", "8080", "Port to serve on")
		devMode   = flag.Bool("dev", false, "Run in development mode (serve from filesystem)")
		staticDir = flag.String("static", "./web", "Static files directory (dev mode only)")
	)
	flag.Parse()

	var fileServer http.Handler

	if *devMode || !hasEmbedded {
		// Development mode: serve from filesystem
		absPath, err := filepath.Abs(*staticDir)
		if err != nil {
			log.Fatalf("Failed to resolve static directory: %v", err)
		}
		
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			log.Fatalf("Static directory does not exist: %s", absPath)
		}
		
		log.Printf("Development mode: serving from %s", absPath)
		fileServer = http.FileServer(http.Dir(absPath))
	} else {
		// Production mode: serve from embedded files
		webFS, err := fs.Sub(webFiles, "web")
		if err != nil {
			log.Fatalf("Failed to create sub filesystem: %v", err)
		}
		log.Println("Production mode: serving from embedded files")
		fileServer = http.FileServer(http.FS(webFS))
	}

	// Wrap the file server with CORS headers for WASM
	handler := addCORSHeaders(fileServer)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("Server starting on http://localhost%s", addr)
	
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// addCORSHeaders adds necessary headers for WASM execution
func addCORSHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// These headers are required for SharedArrayBuffer and WASM
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		
		// Set correct MIME type for WASM files
		if filepath.Ext(r.URL.Path) == ".wasm" {
			w.Header().Set("Content-Type", "application/wasm")
		}
		
		next.ServeHTTP(w, r)
	})
}