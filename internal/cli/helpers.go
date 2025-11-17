package cli

import (
	"fmt"
	"strings"
)

func parseSetFlags(values []string) (map[string]string, error) {
	if len(values) == 0 {
		return map[string]string{}, nil
	}

	out := make(map[string]string, len(values))
	for _, raw := range values {
		if raw == "" {
			continue
		}

		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid --set value %q; must be key=value", raw)
		}
		out[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return out, nil
}
