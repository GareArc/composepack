package values

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// Merge merges layered values using ComposePack semantics: maps merge recursively,
// later scalars/arrays override earlier ones.
func Merge(base map[string]any, overlays ...map[string]any) (map[string]any, error) {
	result := deepCopyMap(base)
	if result == nil {
		result = map[string]any{}
	}

	for _, overlay := range overlays {
		if overlay == nil {
			continue
		}
		mergeMaps(result, overlay)
	}

	return result, nil
}

func mergeMaps(dst map[string]any, src map[string]any) {
	for key, srcVal := range src {
		if existing, ok := dst[key]; ok {
			if merged, ok := mergeValues(existing, srcVal); ok {
				dst[key] = merged
				continue
			}
		}
		dst[key] = deepCopyValue(srcVal)
	}
}

func mergeValues(dstVal, srcVal any) (any, bool) {
	dstMap, dstIsMap := toStringMap(dstVal)
	srcMap, srcIsMap := toStringMap(srcVal)
	if dstIsMap && srcIsMap {
		mergeMaps(dstMap, srcMap)
		return dstMap, true
	}

	// For slices and scalars we simply override with the overlay value.
	return deepCopyValue(srcVal), true
}

func deepCopyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = deepCopyValue(v)
	}
	return out
}

func deepCopySlice(src []any) []any {
	if src == nil {
		return nil
	}
	out := make([]any, len(src))
	for i, v := range src {
		out[i] = deepCopyValue(v)
	}
	return out
}

func deepCopyValue(val any) any {
	switch typed := val.(type) {
	case map[string]any:
		return deepCopyMap(typed)
	case []any:
		return deepCopySlice(typed)
	default:
		return typed
	}
}

func toStringMap(val any) (map[string]any, bool) {
	switch typed := val.(type) {
	case map[string]any:
		return typed, true
	default:
		return nil, false
	}
}

// Validate ensures the provided values conform to the optional JSON schema.
func Validate(schema []byte, vals map[string]any) error {
	if len(schema) == 0 {
		return nil
	}
	if vals == nil {
		vals = map[string]any{}
	}

	schemaLoader := gojsonschema.NewBytesLoader(schema)
	docLoader := gojsonschema.NewGoLoader(vals)
	result, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return fmt.Errorf("validate values: %w", err)
	}

	if result.Valid() {
		return nil
	}

	var sb strings.Builder
	for i, desc := range result.Errors() {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(desc.String())
	}

	return errors.New(sb.String())
}
