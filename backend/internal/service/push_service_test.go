package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBuildBarkURL(t *testing.T) {
	tests := []struct {
		mode, title, message, want string
	}{
		{"message", "", "hello world", "https://push.example.com/bark/key/hello%20world"},
		{"title", "News", "hello", "https://push.example.com/bark/key/News/hello"},
		{"sound", "", "hello", "https://push.example.com/bark/key/hello?sound=minuet"},
		{"call", "", "hello", "https://push.example.com/bark/key/hello?call=1"},
		{"critical", "", "hello", "https://push.example.com/bark/key/hello?level=critical&volume=5"},
	}
	for _, tt := range tests {
		got, err := buildBarkURL("https://push.example.com/bark/key/", tt.mode, tt.title, tt.message)
		if err != nil || got != tt.want {
			t.Fatalf("mode %s: got %q, err %v; want %q", tt.mode, got, err, tt.want)
		}
	}
}

func TestNtfyAccessCommandPlacesFlagsAfterSubcommand(t *testing.T) {
	push := &PushService{
		ntfyCLIPath:    "ntfy",
		ntfyConfigFile: "/etc/ntfy/server.yml",
		ntfyUsername:   "blog",
	}
	executable, args, err := push.ntfyAccessCommand("blog-test")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"access", "--config", "/etc/ntfy/server.yml", "blog", "blog-test", "rw"}
	if executable != "ntfy" || !reflect.DeepEqual(args, want) {
		t.Fatalf("got %q %#v, want ntfy %#v", executable, args, want)
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
