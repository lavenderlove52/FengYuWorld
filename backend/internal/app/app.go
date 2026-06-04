package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"my-blog/backend/internal/config"
	"my-blog/backend/internal/db"
	"my-blog/backend/internal/handler"
	"my-blog/backend/internal/markdown"
	"my-blog/backend/internal/middleware"
	"my-blog/backend/internal/repository"
	"my-blog/backend/internal/service"
	"my-blog/backend/internal/storage"
	"my-blog/backend/internal/util"
	"my-blog/backend/internal/ws"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	DB     *gorm.DB
	Router *gin.Engine
	Server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect db: %w", err)
	}

	jwtManager := util.NewJWTManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessExpires,
		cfg.JWTRefreshExpires,
	)

	userRepo := repository.NewUserRepository(database)
	postRepo := repository.NewPostRepository(database)
	friendRepo := repository.NewFriendshipRepository(database)
	categoryRepo := repository.NewCategoryRepository(database)
	tagRepo := repository.NewTagRepository(database)
	uploadRepo := repository.NewUploadRepository(database)
	pinRepo := repository.NewPinRepository(database)
	emailCodeRepo := repository.NewEmailCodeRepository(database)
	inviteCodeRepo := repository.NewInviteCodeRepository(database)
	refreshTokenRepo := repository.NewRefreshTokenRepository(database)

	mdRenderer := markdown.NewRenderer()
	localStorage := storage.NewLocalStorage(cfg.UploadDir, cfg.PublicUploadBaseURL)

	emailSvc := service.NewEmailService(cfg)
	if !emailSvc.IsConfigured() {
		slog.Warn("SMTP not configured: register email verification will fail until SMTP_* env vars are set")
	} else {
		slog.Info("SMTP configured", "host", cfg.SMTPHost, "port", cfg.SMTPPort, "from", cfg.SMTPFrom)
	}
	captchaSvc := service.NewCaptchaService()
	emailCodeSvc := service.NewEmailCodeService(emailCodeRepo, emailSvc)
	inviteSvc := service.NewInviteService(inviteCodeRepo)

	authSvc := service.NewAuthService(database, userRepo, jwtManager, refreshTokenRepo, cfg.JWTRefreshExpires, emailCodeSvc, inviteSvc)
	userSvc := service.NewUserService(userRepo)
	policy := service.NewPostPolicy(friendRepo, postRepo)
	postSvc := service.NewPostService(postRepo, tagRepo, categoryRepo, friendRepo, policy, mdRenderer)
	friendSvc := service.NewFriendshipService(friendRepo, userRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	tagSvc := service.NewTagService(tagRepo)
	uploadSvc := service.NewUploadService(localStorage, uploadRepo, cfg.MaxImageSizeMB)
	pinSvc := service.NewPinService(pinRepo, postRepo, userRepo, friendRepo, policy)

	userThemeRepo := repository.NewUserThemeRepository(database)
	homeThemeImageRepo := repository.NewHomeThemeImageRepository(database)
	userThemeSvc := service.NewUserThemeService(userThemeRepo, userRepo, homeThemeImageRepo)
	homeThemeImageSvc := service.NewHomeThemeImageService(homeThemeImageRepo, uploadSvc)

	wsTicketStore := service.NewWSTicketStore(60 * time.Second)

	authHandler := handler.NewAuthHandler(authSvc, emailCodeSvc, captchaSvc, wsTicketStore)
	captchaHandler := handler.NewCaptchaHandler(captchaSvc)
	userHandler := handler.NewUserHandler(userSvc, postSvc)
	postHandler := handler.NewPostHandler(postSvc)
	friendHandler := handler.NewFriendshipHandler(friendSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	tagHandler := handler.NewTagHandler(tagSvc)
	uploadHandler := handler.NewUploadHandler(uploadSvc)
	healthHandler := handler.NewHealthHandler()
	pinHandler := handler.NewPinHandler(pinSvc)
	inviteHandler := handler.NewInviteHandler(inviteSvc)
	userThemeHandler := handler.NewUserThemeHandler(userThemeSvc)
	homeThemeImageHandler := handler.NewHomeThemeImageHandler(homeThemeImageSvc)

	authMW := middleware.NewAuthMiddlewareWithDB(jwtManager, database, middleware.NewUserCache())

	hub := ws.NewHub()
	msgRepo := repository.NewMessageRepository(database)
	wsHandler := handler.NewWSHandler(hub, msgRepo, friendRepo, wsTicketStore)
	chatHandler := handler.NewChatHandler(msgRepo, friendRepo)

	router := gin.New()
	router.Use(gin.Logger(), middleware.ErrorRecovery(), middleware.CORS())

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		return nil, err
	}
	router.Static(cfg.PublicUploadBaseURL, cfg.UploadDir)

	registerRoutes(router, authMW, authHandler, captchaHandler, userHandler, postHandler, friendHandler, categoryHandler, tagHandler, uploadHandler, healthHandler, pinHandler, inviteHandler, userThemeHandler, homeThemeImageHandler, wsHandler, chatHandler)

	return &App{
		Config: cfg,
		DB:     database,
		Router: router,
		Server: &http.Server{
			Addr:    ":" + cfg.AppPort,
			Handler: router,
		},
	}, nil
}

func (a *App) Run() error {
	slog.Info("server starting", "port", a.Config.AppPort)
	return a.Server.ListenAndServe()
}

func (a *App) Close() error {
	sqlDB, err := a.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func MigrationsDir() string {
	return filepath.Join("migrations")
}
