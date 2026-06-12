package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/repository"
)

var (
	ErrInvalidPushEndpoint = errors.New("invalid push endpoint")
	ErrPushNotConfigured   = errors.New("push endpoint not configured")
	ErrInvalidPushMessage  = errors.New("invalid push message")
	ErrPushFailed          = errors.New("push request failed")
	ErrNtfyNotConfigured   = errors.New("ntfy is not configured")
	ErrInvalidNtfyTopic    = errors.New("invalid ntfy topic")
	ErrNtfyACLFailed       = errors.New("ntfy ACL update failed")
	ErrPublicPushNotFound  = errors.New("public push endpoint not found")
)

var ntfyTopicSuffixPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

type PushService struct {
	settings            *repository.UserPushSettingRepository
	topics              *repository.UserNtfyTopicRepository
	client              *http.Client
	ntfyClient          *http.Client
	ntfyBaseURL         string
	ntfyInternalURL     string
	ntfyUsername        string
	ntfyPassword        string
	ntfyCLIPath         string
	ntfyAuthFile        string
	ntfyDockerContainer string
}

func NewPushService(settings *repository.UserPushSettingRepository, topics *repository.UserNtfyTopicRepository, ntfyBaseURL, ntfyInternalURL, ntfyUsername, ntfyPassword, ntfyCLIPath, ntfyAuthFile, ntfyDockerContainer string) *PushService {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(ips) == 0 {
				return nil, ErrInvalidPushEndpoint
			}
			for _, resolved := range ips {
				if !isPublicIP(resolved.IP) {
					return nil, ErrInvalidPushEndpoint
				}
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		},
	}
	publicURL := strings.TrimRight(strings.TrimSpace(ntfyBaseURL), "/")
	internalURL := strings.TrimRight(strings.TrimSpace(ntfyInternalURL), "/")
	if internalURL == "" {
		internalURL = publicURL
	}
	return &PushService{
		settings:            settings,
		topics:              topics,
		ntfyBaseURL:         publicURL,
		ntfyInternalURL:     internalURL,
		ntfyUsername:        strings.TrimSpace(ntfyUsername),
		ntfyPassword:        ntfyPassword,
		ntfyCLIPath:         strings.TrimSpace(ntfyCLIPath),
		ntfyAuthFile:        strings.TrimSpace(ntfyAuthFile),
		ntfyDockerContainer: strings.TrimSpace(ntfyDockerContainer),
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		ntfyClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (s *PushService) GetSetting(userID int64) (*domain.UserPushSetting, error) {
	return s.settings.Get(userID)
}

func (s *PushService) ListNtfyTopics(userID int64) ([]domain.UserNtfyTopic, error) {
	return s.topics.List(userID)
}

func (s *PushService) UpdateSetting(userID int64, endpoint string) (*domain.UserPushSetting, error) {
	normalized, err := normalizeBarkEndpoint(endpoint)
	if err != nil {
		return nil, err
	}
	setting, err := s.settings.Get(userID)
	if err != nil {
		return nil, err
	}
	setting.BarkEndpoint = normalized
	if err := s.settings.Upsert(setting); err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *PushService) NtfyBaseURL() string { return s.ntfyBaseURL }

func (s *PushService) CreateNtfyTopic(userID int64, suffix string) (*domain.UserNtfyTopic, error) {
	suffix = strings.TrimSpace(suffix)
	if !ntfyTopicSuffixPattern.MatchString(suffix) {
		return nil, ErrInvalidNtfyTopic
	}
	if s.ntfyBaseURL == "" || s.ntfyInternalURL == "" || s.ntfyUsername == "" {
		return nil, ErrNtfyNotConfigured
	}
	topic := "blog-" + suffix
	if err := s.grantNtfyAccess(topic); err != nil {
		return nil, err
	}
	created := &domain.UserNtfyTopic{UserID: userID, Topic: topic}
	if err := s.topics.Create(created); err != nil {
		return nil, err
	}
	return created, nil
}

func (s *PushService) UpdateNtfyPublic(userID, topicID int64, enabled bool) (*domain.UserNtfyTopic, error) {
	topic, err := s.topics.GetOwned(userID, topicID)
	if err != nil {
		return nil, err
	}
	if topic == nil {
		return nil, ErrPushNotConfigured
	}
	if topic.PublicToken == "" {
		topic.PublicToken, err = newPublicPushToken()
		if err != nil {
			return nil, err
		}
	}
	topic.PublicEnabled = enabled
	if err := s.topics.Update(topic); err != nil {
		return nil, err
	}
	return topic, nil
}

func (s *PushService) SendNtfy(userID, topicID int64, mode, title, message string, priority int, tags string) error {
	topic, err := s.topics.GetOwned(userID, topicID)
	if err != nil {
		return err
	}
	if topic == nil {
		return ErrPushNotConfigured
	}
	return s.sendNtfy(topic.Topic, mode, title, message, priority, tags)
}

func (s *PushService) GetPublicNtfy(token string) (*domain.UserNtfyTopic, error) {
	if len(token) < 32 {
		return nil, ErrPublicPushNotFound
	}
	setting, err := s.topics.GetByPublicToken(token)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, ErrPublicPushNotFound
	}
	return setting, nil
}

func (s *PushService) SendPublicNtfy(token, mode, title, message string, priority int, tags string) error {
	setting, err := s.GetPublicNtfy(token)
	if err != nil {
		return err
	}
	return s.sendNtfy(setting.Topic, mode, title, message, priority, tags)
}

func (s *PushService) sendNtfy(topic, mode, title, message string, priority int, tags string) error {
	if s.ntfyInternalURL == "" || s.ntfyUsername == "" || s.ntfyPassword == "" {
		return ErrNtfyNotConfigured
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return ErrInvalidPushMessage
	}
	payload := map[string]any{"topic": topic, "message": message}
	switch mode {
	case "message":
	case "title":
		title = strings.TrimSpace(title)
		if title == "" {
			return ErrInvalidPushMessage
		}
		payload["title"] = title
	case "urgent":
		payload["priority"] = 5
	case "priority":
		if priority < 1 || priority > 5 {
			return ErrInvalidPushMessage
		}
		payload["priority"] = priority
	case "tags":
		parsedTags := normalizeNtfyTags(tags)
		if len(parsedTags) == 0 {
			parsedTags = []string{"warning", "rotating_light"}
		}
		payload["tags"] = parsedTags
	default:
		return ErrInvalidPushMessage
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, s.ntfyInternalURL, strings.NewReader(string(body)))
	if err != nil {
		return ErrNtfyNotConfigured
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(s.ntfyUsername, s.ntfyPassword)
	resp, err := s.ntfyClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPushFailed, err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: ntfy returned HTTP %d", ErrPushFailed, resp.StatusCode)
	}
	return nil
}

func (s *PushService) grantNtfyAccess(topic string) error {
	executable, args, err := s.ntfyAccessCommand(topic)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = environmentWithout(os.Environ(), "NTFY_CONFIG_FILE")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrNtfyACLFailed, strings.TrimSpace(string(output)))
	}
	return nil
}

func (s *PushService) ntfyAccessCommand(topic string) (string, []string, error) {
	if s.ntfyCLIPath == "" {
		return "", nil, ErrNtfyACLFailed
	}
	args := []string{}
	executable := s.ntfyCLIPath
	if s.ntfyDockerContainer != "" {
		executable = "docker"
		args = append(args, "exec", s.ntfyDockerContainer, s.ntfyCLIPath)
	}
	args = append(args, "access")
	if s.ntfyAuthFile == "" {
		return "", nil, fmt.Errorf("%w: NTFY_AUTH_FILE is required for ntfy access", ErrNtfyACLFailed)
	}
	args = append(args, "--auth-file", s.ntfyAuthFile)
	args = append(args, s.ntfyUsername, topic, "rw")
	return executable, args, nil
}

func environmentWithout(environment []string, keys ...string) []string {
	blocked := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		blocked[key] = struct{}{}
	}
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		key, _, found := strings.Cut(entry, "=")
		if _, remove := blocked[key]; !found || !remove {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func normalizeNtfyTags(raw string) []string {
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" && len(tag) <= 40 {
			tags = append(tags, tag)
		}
		if len(tags) == 5 {
			break
		}
	}
	return tags
}

func newPublicPushToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *PushService) SendBark(userID int64, mode, title, message string) error {
	setting, err := s.settings.Get(userID)
	if err != nil {
		return err
	}
	if setting.BarkEndpoint == "" {
		return ErrPushNotConfigured
	}
	target, err := buildBarkURL(setting.BarkEndpoint, mode, title, message)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return ErrInvalidPushMessage
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPushFailed, err)
	}
	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := barkErrorMessage(responseBody)
		return fmt.Errorf("%w: bark returned HTTP %d: %s", ErrPushFailed, resp.StatusCode, message)
	}
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if len(responseBody) > 0 && json.Unmarshal(responseBody, &result) == nil && result.Code != 0 && result.Code != 200 {
		return fmt.Errorf("%w: bark returned code %d: %s", ErrPushFailed, result.Code, result.Message)
	}
	return nil
}

func normalizeBarkEndpoint(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return "", ErrInvalidPushEndpoint
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		return "", ErrInvalidPushEndpoint
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return "", ErrInvalidPushEndpoint
	}
	if ip := net.ParseIP(host); ip != nil && !isPublicIP(ip) {
		return "", ErrInvalidPushEndpoint
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "" {
		return "", ErrInvalidPushEndpoint
	}
	return parsed.String(), nil
}

func buildBarkURL(endpoint, mode, title, message string) (string, error) {
	base, err := normalizeBarkEndpoint(endpoint)
	if err != nil || strings.TrimSpace(message) == "" {
		return "", ErrInvalidPushMessage
	}
	segments := []string{url.PathEscape(strings.TrimSpace(message))}
	query := url.Values{}
	switch mode {
	case "message":
	case "title":
		if strings.TrimSpace(title) == "" {
			return "", ErrInvalidPushMessage
		}
		segments = []string{url.PathEscape(strings.TrimSpace(title)), segments[0]}
	case "sound":
		query.Set("sound", "alarm")
	case "call":
		query.Set("call", "1")
	case "critical":
		query.Set("level", "critical")
		query.Set("volume", "5")
		query.Set("call", "1")
	default:
		return "", ErrInvalidPushMessage
	}
	target := base + "/" + strings.Join(segments, "/")
	if encoded := query.Encode(); encoded != "" {
		target += "?" + encoded
	}
	return target, nil
}

func barkErrorMessage(body []byte) string {
	var result struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &result) == nil && strings.TrimSpace(result.Message) != "" {
		return strings.TrimSpace(result.Message)
	}
	message := strings.TrimSpace(string(body))
	if message == "" {
		return "empty response"
	}
	return message
}

func isPublicIP(ip net.IP) bool {
	return ip != nil && !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified() && !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsMulticast()
}
