package governance

import (
	"strings"
	"testing"
)

func TestMaskJSONMasksQueryAndContent(t *testing.T) {
	v := map[string]any{
		"query":   "how do I reset my password",
		"content": "secret internal doc body",
		"name":    "profile-1",
	}
	got := maskJSON(v)
	if strings.Contains(got, "reset my password") {
		t.Fatalf("query value should be masked, got %s", got)
	}
	if strings.Contains(got, "secret internal doc body") {
		t.Fatalf("content value should be masked, got %s", got)
	}
	if !strings.Contains(got, "[masked]") {
		t.Fatalf("masked placeholder missing, got %s", got)
	}
	if !strings.Contains(got, `"name":"profile-1"`) {
		t.Fatalf("non-sensitive field should survive, got %s", got)
	}
}

func TestMaskJSONNilIsEmpty(t *testing.T) {
	if got := maskJSON(nil); got != "" {
		t.Fatalf("nil should mask to empty string, got %q", got)
	}
}

func TestMaskIPKeepsFirstTwoOctets(t *testing.T) {
	cases := map[string]string{
		"192.168.1.23": "192.168.*.*",
		"10.0.0.1":     "10.0.*.*",
		"localhost":    "localhost",
		"127.0.0.1":    "127.0.*.*",
	}
	for in, want := range cases {
		if got := maskIP(in); got != want {
			t.Fatalf("maskIP(%q) = %q, want %q", in, got, want)
		}
	}
}
