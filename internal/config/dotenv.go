package config

import (
	"os"
	"strings"
)

// LoadDotEnv loads KEY=VALUE pairs from path into the process environment. The
// real environment wins (an already-set var is not overridden), and empty values
// are ignored so they can't mask a key set elsewhere. Missing file -> no-op.
// Returns the number of vars actually set.
func LoadDotEnv(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	n := 0
	for k, v := range parseDotEnv(string(data)) {
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
			n++
		}
	}
	return n
}

// parseDotEnv parses .env content: KEY=VALUE lines, # comments, blank lines, an
// optional `export` prefix, and single/double-quoted values. Empty values and
// lines without `=` are skipped.
func parseDotEnv(content string) map[string]string {
	out := map[string]string{}
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, _ := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}
		if key == "" || val == "" {
			continue
		}
		out[key] = val
	}
	return out
}
