package middleware

import (
	"net/http"
	"strings"

	"my-blog/backend/internal/domain"
	"my-blog/backend/internal/dto"
	"my-blog/backend/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ContextUserIDKey = "userID"
const ContextUsernameKey = "username"

type AuthMiddleware struct {
	jwt   *util.JWTManager
	db    *gorm.DB
	cache *UserCache
}

func NewAuthMiddleware(jwt *util.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func NewAuthMiddlewareWithDB(jwt *util.JWTManager, db *gorm.DB, cache *UserCache) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt, db: db, cache: cache}
}

func (m *AuthMiddleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, username, ok := m.parseBearer(c)
		if ok {
			if m.db != nil {
				cu, err := m.lookupUser(uid)
				if err == nil && cu.Status == domain.UserStatusActive {
					c.Set(ContextUserIDKey, uid)
					c.Set(ContextUsernameKey, username)
					c.Set("userRole", string(cu.Role))
					c.Set("userCanSitePin", cu.CanSitePin)
				}
			} else {
				c.Set(ContextUserIDKey, uid)
				c.Set(ContextUsernameKey, username)
			}
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

		if m.db != nil {
			cu, err := m.lookupUser(uid)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Err(dto.CodeUserRevoked, "会话已失效，请重新登录"))
				return
			}
			if cu.Status == domain.UserStatusLocked {
				c.AbortWithStatusJSON(http.StatusForbidden, dto.Err(dto.CodeAccountLocked, "账号已被锁定"))
				return
			}
			c.Set("userRole", string(cu.Role))
			c.Set("userCanSitePin", cu.CanSitePin)
		}

		c.Next()
	}
}

func (m *AuthMiddleware) lookupUser(uid int64) (cachedUser, error) {
	if m.cache != nil {
		if cu, ok := m.cache.Get(uid); ok {
			return cu, nil
		}
	}
	var user domain.User
	if err := m.db.Select("id, role, status, can_site_pin").First(&user, uid).Error; err != nil {
		return cachedUser{}, err
	}
	cu := cachedUser{
		Status:     user.Status,
		Role:       user.Role,
		CanSitePin: user.CanSitePin,
	}
	if m.cache != nil {
		m.cache.Set(uid, cu)
	}
	return cu, nil
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

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, username, ok := m.parseBearer(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Err(dto.CodeUnauthorized, "未登录"))
			return
		}
		c.Set(ContextUserIDKey, uid)
		c.Set(ContextUsernameKey, username)

		if m.db == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Err(dto.CodeInternalError, "系统错误"))
			return
		}
		var user domain.User
		if err := m.db.Select("id, role").First(&user, uid).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "权限不足"))
			return
		}
		if user.Role != domain.UserRoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.Err(dto.CodeForbidden, "需要管理员权限"))
			return
		}
		c.Next()
	}
}
