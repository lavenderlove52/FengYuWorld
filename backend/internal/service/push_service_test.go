package service

import (
	"encoding/json"
	"errors"
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
		{"sound", "", "hello", "https://push.example.com/bark/key/hello?sound=default"},
		{"call", "", "hello", "https://push.example.com/bark/key/hello?call=1"},
		{"critical", "", "hello", "https://push.example.com/bark/key/hello?call=1&level=critical&sound=default&volume=5"},
	}
	for _, tt := range tests {
		got, err := buildBarkURL("https://push.example.com/bark/key/", tt.mode, tt.title, tt.message)
		if err != nil || got != tt.want {
			t.Fatalf("mode %s: got %q, err %v; want %q", tt.mode, got, err, tt.want)
		}
	}
}

func TestBuildBarkRequestUsesV2PushPayload(t *testing.T) {
	target, body, err := buildBarkRequest("https://push.example.com/bark/device-key", "sound", "", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if target != "https://push.example.com/bark/push" {
		t.Fatalf("got target %q", target)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["device_key"] != "device-key" || payload["body"] != "hello" || payload["sound"] != "default" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestNtfyAccessCommandPlacesFlagsAfterSubcommand(t *testing.T) {
	push := &PushService{
		ntfyCLIPath:  "ntfy",
		ntfyAuthFile: "/var/lib/ntfy/user.db",
		ntfyUsername: "blog",
	}
	executable, args, err := push.ntfyAccessCommand("blog-test")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"access", "--auth-file", "/var/lib/ntfy/user.db", "blog", "blog-test", "rw"}
	if executable != "ntfy" || !reflect.DeepEqual(args, want) {
		t.Fatalf("got %q %#v, want ntfy %#v", executable, args, want)
	}
}

func TestNtfyAccessCommandRejectsConfigOnly(t *testing.T) {
	push := &PushService{ntfyCLIPath: "ntfy", ntfyUsername: "blog"}
	if _, _, err := push.ntfyAccessCommand("blog-test"); !errors.Is(err, ErrNtfyACLFailed) {
		t.Fatalf("expected ErrNtfyACLFailed, got %v", err)
	}
}

func TestNtfyAccessCommandPrefersAuthFile(t *testing.T) {
	push := &PushService{
		ntfyCLIPath:  "ntfy",
		ntfyAuthFile: "/var/lib/ntfy/user.db",
		ntfyUsername: "blog",
	}
	_, args, err := push.ntfyAccessCommand("blog-test")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"access", "--auth-file", "/var/lib/ntfy/user.db", "blog", "blog-test", "rw"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("got %#v, want %#v", args, want)
	}
}

func TestEnvironmentWithoutNtfyConfigFile(t *testing.T) {
	got := environmentWithout([]string{"PATH=/usr/bin", "NTFY_CONFIG_FILE=/missing.yml", "HOME=/root"}, "NTFY_CONFIG_FILE")
	want := []string{"PATH=/usr/bin", "HOME=/root"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
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

func TestSendNtfyRequiresPassword(t *testing.T) {
	push := &PushService{ntfyInternalURL: "http://127.0.0.1:10002", ntfyUsername: "blog", ntfyClient: http.DefaultClient}
	if !errors.Is(push.sendNtfy("blog-phone", "message", "", "hello", 0, ""), ErrNtfyNotConfigured) {
		t.Fatal("expected missing password to reject ntfy send")
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
