package config

import (
	"testing"

	. "github.com/infrago/base"
)

func TestRedact(t *testing.T) {
	in := Map{
		"username": "admin",
		"password": "secret",
		"setting": Map{
			"api_key": "token",
			"addr":    "127.0.0.1:6379",
		},
		"items": []Any{
			Map{"token": "abc", "name": "demo"},
		},
	}

	out := Redact(in)
	if out["username"] != "admin" {
		t.Fatalf("username=%v, want admin", out["username"])
	}
	if out["password"] != redactedValue {
		t.Fatalf("password=%v, want redacted", out["password"])
	}
	nested := out["setting"].(Map)
	if nested["api_key"] != redactedValue {
		t.Fatalf("api_key=%v, want redacted", nested["api_key"])
	}
	items := out["items"].([]Any)
	item := items[0].(Map)
	if item["token"] != redactedValue {
		t.Fatalf("token=%v, want redacted", item["token"])
	}
	if in["password"] != "secret" {
		t.Fatalf("input mutated: %v", in["password"])
	}
}
