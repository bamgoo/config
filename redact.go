package config

import (
	"strings"

	. "github.com/infrago/base"
)

const redactedValue = "***"

func Redact(value Map) Map {
	out, _ := redactAny(value).(Map)
	return out
}

func redactAny(value Any) Any {
	switch v := value.(type) {
	case Map:
		out := Map{}
		for key, val := range v {
			if isSensitiveKey(key) {
				out[key] = redactedValue
				continue
			}
			out[key] = redactAny(val)
		}
		return out
	case []Any:
		out := make([]Any, len(v))
		for i, item := range v {
			out[i] = redactAny(item)
		}
		return out
	case []Map:
		out := make([]Map, len(v))
		for i, item := range v {
			out[i] = Redact(item)
		}
		return out
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, ".", "_")

	switch key {
	case "password", "passwd", "pwd", "secret", "token", "key", "api_key", "apikey", "access_key", "private_key", "credential", "credentials":
		return true
	}

	for _, part := range strings.Split(key, "_") {
		switch part {
		case "password", "passwd", "pwd", "secret", "token":
			return true
		}
	}
	return false
}
