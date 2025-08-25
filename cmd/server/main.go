package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mbarlow/local-first/internal/monitoring"
)

func main() {
	var (
		port      = flag.String("port", "8080", "Port to serve on")
		devMode   = flag.Bool("dev", false, "Run in development mode (serve from filesystem)")
		staticDir = flag.String("static", "./web", "Static files directory (dev mode only)")
	)
	flag.Parse()

	var fileServer http.Handler

	if *devMode {
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
		if !hasEmbedded {
			log.Fatal("No embedded files available. Build with -tags embed or use -dev flag")
		}
		webFS, err := fs.Sub(webFiles, "web")
		if err != nil {
			log.Fatalf("Failed to create sub filesystem: %v", err)
		}
		log.Println("Production mode: serving from embedded files")
		fileServer = http.FileServer(http.FS(webFS))
	}

	// Add monitoring middleware
	monitor := monitoring.NewMonitor()
	
	// Wrap the file server with CORS headers for WASM
	corsHandler := addCORSHeaders(fileServer)
	
	// Add monitoring
	handler := monitor.Middleware(corsHandler)

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