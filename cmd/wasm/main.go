//go:build wasm
// +build wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/mbarlow/local-first/internal/api"
)

func main() {
	fmt.Println("Go WASM API loaded successfully!")

	// Create API handler instance
	apiHandler := api.NewHandler()

	// Create a JavaScript object to hold our API functions
	goAPI := js.Global().Get("Object").New()
	
	// Register each function individually on the goAPI object
	goAPI.Set("processData", js.FuncOf(apiHandler.ProcessData))
	goAPI.Set("validateInput", js.FuncOf(apiHandler.ValidateInput))
	goAPI.Set("calculateStats", js.FuncOf(apiHandler.CalculateStats))
	goAPI.Set("formatJSON", js.FuncOf(apiHandler.FormatJSON))
	goAPI.Set("generateID", js.FuncOf(apiHandler.GenerateID))
	goAPI.Set("getVersion", js.FuncOf(apiHandler.GetVersion))
	
	// Add a simple test function
	goAPI.Set("test", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		fmt.Println("Test function called")
		result := js.Global().Get("Object").New()
		result.Set("success", true)
		result.Set("message", "Test successful")
		result.Set("data", "hello world")
		return result
	}))
	
	// Set the goAPI object on the global window
	js.Global().Set("goAPI", goAPI)

	// Register cleanup callback
	js.Global().Set("goAPICleanup", js.FuncOf(cleanup))

	fmt.Println("Go API functions registered globally as 'goAPI'")
	fmt.Println("Available functions: processData, validateInput, calculateStats, formatJSON, generateID, getVersion")

	// Keep the Go program alive
	<-make(chan bool)
}

// cleanup releases Go resources when called from JavaScript
func cleanup(this js.Value, inputs []js.Value) interface{} {
	fmt.Println("Cleaning up Go WASM resources...")
	return map[string]interface{}{
		"success": true,
		"message": "Cleanup complete",
	}
}

// Helper function to safely convert JS values to Go types
func jsValueToInterface(val js.Value) interface{} {
	switch val.Type() {
	case js.TypeString:
		return val.String()
	case js.TypeNumber:
		return val.Float()
	case js.TypeBoolean:
		return val.Bool()
	case js.TypeObject:
		if val.Get("constructor").Get("name").String() == "Array" {
			length := val.Get("length").Int()
			slice := make([]interface{}, length)
			for i := 0; i < length; i++ {
				slice[i] = jsValueToInterface(val.Index(i))
			}
			return slice
		}
		// Handle objects by converting to map
		obj := make(map[string]interface{})
		// Note: In a real implementation, you'd need to iterate over object properties
		// This is simplified for the example
		return obj
	default:
		return nil
	}
}

// Helper function to create standardized API responses
func createAPIResponse(success bool, data interface{}, message string) map[string]interface{} {
	return map[string]interface{}{
		"success":   success,
		"data":      data,
		"message":   message,
		"timestamp": js.Global().Get("Date").New().Call("toISOString").String(),
	}
}
