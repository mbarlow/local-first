import { defineConfig } from 'vite'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

export default defineConfig({
  root: 'web',
  server: {
    port: 5173,
    headers: {
      'Cross-Origin-Embedder-Policy': 'require-corp',
      'Cross-Origin-Opener-Policy': 'same-origin'
    },
    fs: {
      // Allow serving files from the web directory
      strict: false
    }
  },
  build: {
    outDir: '../dist',
    emptyOutDir: true
  },
  optimizeDeps: {
    exclude: ['wasm_exec.js']
  },
  assetsInclude: ['**/*.wasm']
})