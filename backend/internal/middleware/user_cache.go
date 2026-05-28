package middleware

import (
	"sync"
	"time"

	"my-blog/backend/internal/domain"
)

type cachedUser struct {
	Status     domain.UserStatus
	Role       domain.UserRole
	CanSitePin bool
	ExpiresAt  time.Time
}

type UserCache struct {
	mu    sync.RWMutex
	items map[int64]cachedUser
}

func NewUserCache() *UserCache {
	return &UserCache{items: make(map[int64]cachedUser)}
}

func (c *UserCache) Get(id int64) (cachedUser, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[id]
	if !ok || time.Now().After(item.ExpiresAt) {
		return cachedUser{}, false
	}
	return item, true
}

func (c *UserCache) Set(id int64, u cachedUser) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u.ExpiresAt = time.Now().Add(30 * time.Second)
	c.items[id] = u
}

func (c *UserCache) Invalidate(id int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, id)
}
