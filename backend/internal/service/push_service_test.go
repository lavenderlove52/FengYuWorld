package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildBarkRequest(t *testing.T) {
	tests := []struct {
		mode, title, message string
		want                 map[string]any
	}{
		{"message", "", "hello world", map[string]any{"device_key": "key", "body": "hello world"}},
		{"title", "News", "hello", map[string]any{"device_key": "key", "body": "hello", "title": "News"}},
		{"sound", "", "hello", map[string]any{"device_key": "key", "body": "hello", "sound": "minuet"}},
		{"call", "", "hello", map[string]any{"device_key": "key", "body": "hello", "call": "1"}},
		{"critical", "", "hello", map[string]any{"device_key": "key", "body": "hello", "level": "critical", "volume": 5}},
	}
	for _, tt := range tests {
		target, payload, err := buildBarkRequest("https://push.example.com/bark/key/", tt.mode, tt.title, tt.message)
		if err != nil || target != "https://push.example.com/bark/push" {
			t.Fatalf("mode %s: target %q, err %v", tt.mode, target, err)
		}
		got, _ := json.Marshal(payload)
		want, _ := json.Marshal(tt.want)
		if string(got) != string(want) {
			t.Fatalf("mode %s: payload %s, want %s", tt.mode, got, want)
		}
	}
}

func TestSendNtfyUrgentWithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "blog" || password != "secret" {
			t.Fatalf("unexpected auth")
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["topic"] != "blog-phone" || payload["priority"] != float64(5) {
			t.Fatalf("unexpected payload: %#v", payload)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	push := &PushService{ntfyBaseURL: server.URL, ntfyInternalURL: server.URL, ntfyUsername: "blog", ntfyPassword: "secret", ntfyClient: server.Client()}
	if err := push.sendNtfy("blog-phone", "urgent", "", "wake up", 0, ""); err != nil {
		t.Fatal(err)
	}
}

func TestNormalizeNtfyTags(t *testing.T) {
	tags := normalizeNtfyTags(" warning, rotating_light, ,computer ")
	if len(tags) != 3 || tags[0] != "warning" || tags[2] != "computer" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}

func TestNormalizeBarkEndpointRejectsPrivateTargets(t *testing.T) {
	for _, endpoint := range []string{"http://localhost/bark/key", "http://127.0.0.1/bark/key", "http://192.168.1.2/bark/key"} {
		if _, err := normalizeBarkEndpoint(endpoint); err == nil {
			t.Fatalf("expected %q to be rejected", endpoint)
		}
	}
}
