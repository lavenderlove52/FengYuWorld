package app

import (
	"my-blog/backend/internal/handler"
	"my-blog/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerRoutes(
	r *gin.Engine,
	authMW *middleware.AuthMiddleware,
	auth *handler.AuthHandler,
	users *handler.UserHandler,
	posts *handler.PostHandler,
	friends *handler.FriendshipHandler,
	categories *handler.CategoryHandler,
	tags *handler.TagHandler,
	uploads *handler.UploadHandler,
	health *handler.HealthHandler,
	pins *handler.PinHandler,
	invite *handler.InviteHandler,
	theme *handler.UserThemeHandler,
) {
	api := r.Group("/api")
	{
		api.GET("/health", health.Health)

		authGroup := api.Group("/auth")
		{
			authGroup.POST("/send-email-code", auth.SendEmailCode)
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/refresh", auth.Refresh)
			authGroup.POST("/logout", auth.Logout)
		}

		api.GET("/posts/public", posts.ListPublic)
		api.GET("/posts/site-pins", pins.GetSitePins)
		api.GET("/posts/:id", authMW.Optional(), posts.GetDetail)

		api.GET("/users/:username", users.GetUserProfile)
		api.GET("/users/:username/posts", authMW.Optional(), users.GetUserPosts)
		api.GET("/users/:username/pins", authMW.Optional(), pins.GetPublicUserPins)
		api.GET("/users/:username/theme", theme.GetTheme)

		authed := api.Group("")
		authed.Use(authMW.Required())
		{
			authed.GET("/users/me", users.GetMe)
			authed.PUT("/users/me", users.UpdateMe)
			authed.PUT("/me/theme", theme.UpdateTheme)

			authed.GET("/me/posts", posts.ListMine)
			authed.POST("/posts", posts.Create)
			authed.PUT("/posts/:id", posts.Update)
			authed.DELETE("/posts/:id", posts.Delete)

			authed.POST("/friends/requests", friends.SendRequest)
			authed.GET("/friends/requests/received", friends.ListReceived)
			authed.GET("/friends/requests/sent", friends.ListSent)
			authed.POST("/friends/requests/:id/accept", friends.Accept)
			authed.POST("/friends/requests/:id/reject", friends.Reject)
			authed.GET("/friends", friends.ListFriends)
			authed.DELETE("/friends/:user_id", friends.DeleteFriend)

			authed.GET("/categories", categories.List)
			authed.POST("/categories", categories.Create)
			authed.PUT("/categories/:id", categories.Update)
			authed.DELETE("/categories/:id", categories.Delete)

			authed.GET("/tags", tags.List)
			authed.POST("/tags", tags.Create)
			authed.DELETE("/tags/:id", tags.Delete)

			authed.POST("/uploads/image", uploads.UploadImage)
			authed.POST("/uploads/avatar", uploads.UploadAvatar)

			authed.POST("/pins/site/:post_id", pins.SitePinPost)
			authed.DELETE("/pins/site/:post_id", pins.SiteUnpinPost)
			authed.GET("/me/pins", pins.GetMyPins)
			authed.POST("/me/posts/:post_id/pin", pins.UserPinPost)
			authed.DELETE("/me/posts/:post_id/pin", pins.UserUnpinPost)
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(authMW.RequireAdmin())
		{
			adminGroup.GET("/users", users.ListUsers)
			adminGroup.POST("/users/:user_id/site-pin", pins.GrantSitePinPermission)
			adminGroup.DELETE("/users/:user_id/site-pin", pins.RevokeSitePinPermission)
			adminGroup.POST("/users/:user_id/custom-js", users.GrantCustomJS)
			adminGroup.DELETE("/users/:user_id/custom-js", users.RevokeCustomJS)

			adminGroup.GET("/invite-codes", invite.ListInviteCodes)
			adminGroup.POST("/invite-codes", invite.GenerateInviteCodes)
			adminGroup.DELETE("/invite-codes/:id", invite.DeleteInviteCode)
		}
	}
}
