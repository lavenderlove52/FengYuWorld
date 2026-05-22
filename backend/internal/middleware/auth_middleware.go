package middleware

import (
	"net/http"
	"strings"

	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/util"

	"github.com/gin-gonic/gin"
)

const ContextUserIDKey = "userID"
const ContextUsernameKey = "username"

type AuthMiddleware struct {
	jwt *util.JWTManager
}

func NewAuthMiddleware(jwt *util.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid, username, ok := m.parseBearer(c); ok {
			c.Set(ContextUserIDKey, uid)
			c.Set(ContextUsernameKey, username)
		}
		c.Next()
	}
}

func (m *AuthMiddleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, username, ok := m.parseBearer(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Err(dto.CodeUnauthorized, "未登录"))
			return
		}
		c.Set(ContextUserIDKey, uid)
		c.Set(ContextUsernameKey, username)
		c.Next()
	}
}

func (m *AuthMiddleware) parseBearer(c *gin.Context) (int64, string, bool) {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return 0, "", false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	claims, err := m.jwt.ParseAccessToken(token)
	if err != nil {
		return 0, "", false
	}
	return claims.UserID, claims.Username, true
}

func GetUserID(c *gin.Context) (*int64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return nil, false
	}
	id := v.(int64)
	return &id, true
}

func MustUserID(c *gin.Context) int64 {
	id, ok := GetUserID(c)
	if !ok {
		return 0
	}
	return *id
}
