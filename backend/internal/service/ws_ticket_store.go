package service

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type WSTicketStore struct {
	mu      sync.Mutex
	ttl     time.Duration
	tickets map[string]wsTicket
}

type wsTicket struct {
	UserID    int64
	ExpiresAt time.Time
}

func NewWSTicketStore(ttl time.Duration) *WSTicketStore {
	return &WSTicketStore{
		ttl:     ttl,
		tickets: make(map[string]wsTicket),
	}
}

func (s *WSTicketStore) Issue(userID int64) (string, time.Time, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", time.Time{}, err
	}
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	expiresAt := time.Now().Add(s.ttl)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[ticket] = wsTicket{UserID: userID, ExpiresAt: expiresAt}
	s.deleteExpiredLocked(time.Now())
	return ticket, expiresAt, nil
}

func (s *WSTicketStore) Consume(ticket string) (int64, bool) {
	if ticket == "" {
		return 0, false
	}
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.tickets[ticket]
	delete(s.tickets, ticket)
	if !ok || !item.ExpiresAt.After(now) {
		return 0, false
	}
	s.deleteExpiredLocked(now)
	return item.UserID, true
}

func (s *WSTicketStore) deleteExpiredLocked(now time.Time) {
	for ticket, item := range s.tickets {
		if !item.ExpiresAt.After(now) {
			delete(s.tickets, ticket)
		}
	}
}
