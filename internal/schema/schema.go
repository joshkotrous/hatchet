package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
)

// Define constants for input validation and security limits
const (
	MaxJSONSize       = 10 * 1024 * 1024 // 10MB max JSON size
	MaxArrayLength    = 10000            // Maximum array length
	MaxObjectDepth    = 100              // Maximum nesting depth
	MaxObjectElements = 1000             // Maximum number of elements in an object
	MaxParsingDepth   = 20              // Maximum parsing depth
	MaxObjectFields   = 100             // Maximum fields in an object
	MaxKeyLength      = 256             // Maximum length of a key
)

// Security errors
var (
	ErrJSONTooLarge     = errors.New("JSON input too large")
	ErrJSONTooComplex   = errors.New("JSON structure too complex")
	ErrMaxDepthExceeded = errors.New("maximum parsing depth exceeded")
	ErrTooManyFields    = errors.New("object contains too many fields")
	ErrArrayTooLarge    = errors.New("array contains too many elements")
)

func SchemaBytesFromBytes(d []byte) ([]byte, error) {
	// Check JSON size
	if len(d) > MaxJSONSize {
		return nil, ErrJSONTooLarge
	}

	var m map[string]interface{}
	if err := json.Unmarshal(d, &m); err != nil {
		return nil, err
	}

	// Validate JSON structure
	if err := validateJSONStructure(m, 0); err != nil {
		return nil, err
	}

	return SchemaBytesFromMap(m)
}

func SchemaBytesFromMap(m map[string]interface{}) ([]byte, error) {
	// Use secure parsing with depth tracking
	goType, err := parseWithDepth(m, 0)
	if err != nil {
		return nil, fmt.Errorf("schema generation failed: %w", err)
	}

	// create instance of reflect type
	t := reflect.New(goType).Elem()

	s := jsonschema.Reflect(t.Interface())

	return json.Marshal(s)
}

// validateJSONStructure checks JSON for excessive nesting or size
func validateJSONStructure(data interface{}, depth int) error {
	// Check for excessive nesting
	if depth > MaxObjectDepth {
		return ErrJSONTooComplex
	}

	switch v := data.(type) {
	case map[string]interface{}:
		// Check for excessive object size
		if len(v) > MaxObjectElements {
			return ErrJSONTooComplex
		}
		
		// Recursively check each element
		for _, val := range v {
			if err := validateJSONStructure(val, depth+1); err != nil {
				return err
			}
		}
	case []interface{}:
		// Check for excessive array length
		if len(v) > MaxArrayLength {
			return ErrJSONTooComplex
		}
		
		// Recursively check each element
		for _, val := range v {
			if err := validateJSONStructure(val, depth+1); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// parseWithDepth is a secure version of parse that tracks depth
func parseWithDepth(data interface{}, depth int) (reflect.Type, error) {
	// Check maximum recursion depth
	if depth > MaxParsingDepth {
		return nil, ErrMaxDepthExceeded
	}

	switch v := data.(type) {
	case map[string]interface{}:
		return parseObjectWithDepth(v, depth+1)
	case []interface{}:
		return parseArrayWithDepth(v, depth+1)
	case string:
		return reflect.TypeOf(""), nil
	case float64:
		// Check if it can be an int.
		if v == float64(int(v)) {
			return reflect.TypeOf(0), nil
		}
		return reflect.TypeOf(0.0), nil
	case bool:
		return reflect.TypeOf(false), nil
	case nil:
		return reflect.TypeOf(new(interface{})).Elem(), nil
	default:
		fmt.Printf("Unhandled type: %T\n", v)
		return reflect.TypeOf(new(interface{})).Elem(), nil
	}
}

// parseObjectWithDepth is a secure version of parseObject
func parseObjectWithDepth(obj map[string]interface{}, depth int) (reflect.Type, error) {
	var fields []reflect.StructField
	count := 0

	// Apply a limit to prevent resource exhaustion
	if len(obj) > MaxObjectFields {
		fmt.Printf("Warning: Object exceeds maximum field count: %d > %d\n", len(obj), MaxObjectFields)
		return reflect.TypeOf(struct{}{}), nil // Return a safe default
	}

	for key, val := range obj {
		// Validate key length to prevent DoS
		if len(key) > MaxKeyLength {
			fmt.Printf("Warning: Key exceeds maximum length: %s (%d > %d)\n", key, len(key), MaxKeyLength)
			continue // Skip this field but continue processing others
		}

		fieldType, err := parseWithDepth(val, depth)
		if err != nil {
			return nil, err
		}
		
		// Safely escape key for use in struct tag to prevent injection
		safeKey := strings.Replace(key, `"`, `\"`, -1)
		
		defaultValue := formatDefaultValue(val)
		var tag string
		
		if defaultValue == "" {
			tag = fmt.Sprintf(`json:"%s"`, safeKey)
		} else {
			// Safely escape default value for struct tag
			safeDefaultValue := strings.Replace(defaultValue, `"`, `\"`, -1)
			tag = fmt.Sprintf(`json:"%s" jsonschema:"default=%s"`, safeKey, safeDefaultValue)
		}

		field := reflect.StructField{
			Name: fmt.Sprintf("Field%d", count),
			Type: fieldType,
			Tag:  reflect.StructTag(tag),
		}
		
		fields = append(fields, field)
		count++
	}

	// If all fields were skipped due to validation, return a safe default
	if len(fields) == 0 && len(obj) > 0 {
		fmt.Printf("Warning: All fields were filtered out due to validation\n")
		return reflect.TypeOf(struct{}{}), nil
	}

	return reflect.StructOf(fields), nil
}

// parseArrayWithDepth is a secure version of parseArray
func parseArrayWithDepth(arr []interface{}, depth int) (reflect.Type, error) {
	// Check array length
	if len(arr) > MaxArrayLength {
		return nil, ErrArrayTooLarge
	}

	if len(arr) == 0 {
		return reflect.SliceOf(reflect.TypeOf("")), nil
	}
	
	elemType, err := parseWithDepth(arr[0], depth)
	if err != nil {
		return nil, err
	}
	
	return reflect.SliceOf(elemType), nil
}

func formatDefaultValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case float64, bool:
		return fmt.Sprintf("%v", v)
	case nil:
		return "null"
	default:
		return ""
	}
}

// Legacy functions kept for backward compatibility
func parse(data interface{}) reflect.Type {
	t, _ := parseWithDepth(data, 0)
	return t
}

func parseObject(obj map[string]interface{}) reflect.Type {
	t, _ := parseObjectWithDepth(obj, 0)
	return t
}

func parseArray(arr []interface{}) reflect.Type {
	t, _ := parseArrayWithDepth(arr, 0)
	return t
}