// Recursively projects log fields into bounded JSON-safe redacted values.
// Errors, protected keys, byte payloads and excessive/cyclic nesting are hidden.
package logging

import (
	"reflect"
	"strings"
)

func sensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, word := range []string{"token", "private", "passphrase", "password", "prompt", "model_output", "voice", "audio", "secret", "message_body", "file_content", "payload", "error"} {
		if strings.Contains(lower, word) {
			return true
		}
	}
	return false
}
func Redact(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	result := make(map[string]any, len(fields))
	for key, value := range fields {
		if sensitive(key) {
			result[key] = "[redacted]"
		} else {
			result[key] = redactValue(reflect.ValueOf(value), 0)
		}
	}
	return result
}
func redactValue(value reflect.Value, depth int) any {
	if depth > 8 {
		return "[redacted]"
	}
	if !value.IsValid() {
		return nil
	}
	if value.CanInterface() {
		_, ok := value.Interface().(error)
		if ok {
			return "[redacted]"
		}
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		return redactValue(value.Elem(), depth+1)
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return "[redacted]"
		}
		result := map[string]any{}
		iterator := value.MapRange()
		for iterator.Next() {
			key := iterator.Key().String()
			if sensitive(key) {
				result[key] = "[redacted]"
			} else {
				result[key] = redactValue(iterator.Value(), depth+1)
			}
		}
		return result
	case reflect.Struct:
		result := map[string]any{}
		for index := 0; index < value.NumField(); index++ {
			field := value.Type().Field(index)
			if field.PkgPath != "" {
				continue
			}
			key := strings.Split(field.Tag.Get("json"), ",")[0]
			if key == "-" {
				continue
			}
			if key == "" {
				key = field.Name
			}
			if sensitive(key) || sensitive(field.Name) {
				result[key] = "[redacted]"
			} else {
				result[key] = redactValue(value.Field(index), depth+1)
			}
		}
		return result
	case reflect.Slice, reflect.Array:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return "[redacted]"
		}
		result := make([]any, value.Len())
		for index := range result {
			result[index] = redactValue(value.Index(index), depth+1)
		}
		return result
	case reflect.String:
		return value.String()
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	default:
		return "[redacted]"
	}
}
