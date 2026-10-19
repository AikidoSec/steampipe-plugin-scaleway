package scaleway

import "testing"

func TestRedactURLUserInfo(t *testing.T) {
	tests := map[string]string{
		"user:pass@host:4222":           "host:4222",
		"tok@host":                      "host",
		"1.2.3.4:4222":                  "1.2.3.4:4222",
		"nats://h:4222":                 "nats://h:4222",
		"nats://user:pass@h:4222":       "nats://h:4222",
		"https://u:p@sqs.example.com/q": "https://sqs.example.com/q",
		"host":                          "host",
		"":                              "",
	}
	for in, want := range tests {
		if got := redactURLUserInfo(in); got != want {
			t.Errorf("redactURLUserInfo(%q) = %q, want %q", in, got, want)
		}
	}
}
