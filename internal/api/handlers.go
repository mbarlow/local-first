package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"syscall/js"
	"time"

	"github.com/mbarlow/local-first/internal/core"
)

// Handler contains all API endpoint handlers
type Handler struct {
	processor *core.DataProcessor
}

// NewHandler creates a new API handler instance
func NewHandler() *Handler {
	return &Handler{
		processor: core.NewDataProcessor(),
	}
}

// ProcessData handles data processing requests
func (h *Handler) ProcessData(this js.Value, inputs []js.Value) interface{} {
	fmt.Println("ProcessData called with", len(inputs), "inputs")
	
	if len(inputs) == 0 {
		fmt.Println("No input provided")
		return h.errorResponse("No input provided")
	}

	inputData := inputs[0].String()
	fmt.Printf("Processing input: %s\n", inputData)

	// For now, return a simple response to test
	simpleResult := map[string]interface{}{
		"wordCount": 2,
		"input": inputData,
	}

	return h.successResponse(simpleResult, "Data processed successfully")
}

// ValidateInput validates input data against common patterns
func (h *Handler) ValidateInput(this js.Value, inputs []js.Value) interface{} {
	if len(inputs) < 2 {
		return h.errorResponse("Requires input data and validation type")
	}

	input := inputs[0].String()
	validationType := inputs[1].String()

	isValid, message := h.validateByType(input, validationType)

	return h.successResponse(map[string]interface{}{
		"valid":   isValid,
		"message": message,
		"input":   input,
		"type":    validationType,
	}, "Validation complete")
}

// CalculateStats calculates statistics for numeric arrays
func (h *Handler) CalculateStats(this js.Value, inputs []js.Value) interface{} {
	if len(inputs) == 0 {
		return h.errorResponse("No data provided")
	}

	// Convert JS array to Go slice
	jsArray := inputs[0]
	if jsArray.Type() != js.TypeObject || jsArray.Get("constructor").Get("name").String() != "Array" {
		return h.errorResponse("Input must be an array")
	}

	length := jsArray.Get("length").Int()
	numbers := make([]float64, 0, length)

	for i := 0; i < length; i++ {
		val := jsArray.Index(i)
		if val.Type() == js.TypeNumber {
			numbers = append(numbers, val.Float())
		}
	}

	if len(numbers) == 0 {
		return h.errorResponse("No valid numbers found in array")
	}

	stats := h.processor.CalculateStatistics(numbers)

	return h.successResponse(stats, fmt.Sprintf("Statistics calculated for %d numbers", len(numbers)))
}

// FormatJSON formats and validates JSON strings
func (h *Handler) FormatJSON(this js.Value, inputs []js.Value) interface{} {
	if len(inputs) == 0 {
		return h.errorResponse("No JSON string provided")
	}

	jsonStr := inputs[0].String()

	// Parse and re-format JSON
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid JSON: %v", err))
	}

	formatted, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to format JSON: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"formatted": string(formatted),
		"valid":     true,
		"size":      len(formatted),
	}, "JSON formatted successfully")
}

// GenerateID generates various types of IDs
func (h *Handler) GenerateID(this js.Value, inputs []js.Value) interface{} {
	idType := "uuid"
	if len(inputs) > 0 {
		idType = inputs[0].String()
	}

	id := h.processor.GenerateID(idType)

	return h.successResponse(map[string]interface{}{
		"id":   id,
		"type": idType,
	}, fmt.Sprintf("Generated %s ID", idType))
}

// GetVersion returns API version information
func (h *Handler) GetVersion(this js.Value, inputs []js.Value) interface{} {
	return h.successResponse(map[string]interface{}{
		"version":     "1.0.0",
		"name":        "Go WASM API",
		"buildTime":   time.Now().Format(time.RFC3339),
		"goVersion":   "1.21+",
		"environment": "WebAssembly",
	}, "Version information retrieved")
}

// Helper methods

func (h *Handler) validateByType(input, validationType string) (bool, string) {
	switch validationType {
	case "email":
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if emailRegex.MatchString(input) {
			return true, "Valid email address"
		}
		return false, "Invalid email format"

	case "url":
		if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
			return true, "Valid URL"
		}
		return false, "URL must start with http:// or https://"

	case "phone":
		phoneRegex := regexp.MustCompile(`^\+?[\d\s\-\(\)]{10,}$`)
		if phoneRegex.MatchString(input) {
			return true, "Valid phone number format"
		}
		return false, "Invalid phone number format"

	case "json":
		var obj interface{}
		if json.Unmarshal([]byte(input), &obj) == nil {
			return true, "Valid JSON"
		}
		return false, "Invalid JSON format"

	default:
		return false, fmt.Sprintf("Unknown validation type: %s", validationType)
	}
}

// toJSValue converts a Go value to a JavaScript value recursively
func toJSValue(v interface{}) js.Value {
	if v == nil {
		return js.Null()
	}
	
	switch val := v.(type) {
	case js.Value:
		return val
	case bool:
		return js.ValueOf(val)
	case int:
		return js.ValueOf(val)
	case int64:
		return js.ValueOf(float64(val)) // JS doesn't have int64
	case float64:
		return js.ValueOf(val)
	case string:
		return js.ValueOf(val)
	case []interface{}:
		// Convert slice to JS array
		jsArray := js.Global().Get("Array").New(len(val))
		for i, item := range val {
			jsArray.SetIndex(i, toJSValue(item))
		}
		return jsArray
	case map[string]interface{}:
		// Convert map to JS object
		jsObj := js.Global().Get("Object").New()
		for key, value := range val {
			jsObj.Set(key, toJSValue(value))
		}
		return jsObj
	default:
		// Try to convert using js.ValueOf as fallback
		return js.ValueOf(val)
	}
}

func (h *Handler) successResponse(data interface{}, message string) js.Value {
	response := map[string]interface{}{
		"success":   true,
		"data":      data,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}
	
	return toJSValue(response)
}

func (h *Handler) errorResponse(message string) js.Value {
	response := map[string]interface{}{
		"success":   false,
		"error":     message,
		"timestamp": time.Now().Unix(),
	}
	
	return toJSValue(response)
}
