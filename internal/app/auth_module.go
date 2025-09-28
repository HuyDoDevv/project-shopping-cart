package app

import (
	v1handler "gin/user-management-api/internal/handler/v1"
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/routes"
	v1routes "gin/user-management-api/internal/routes/v1"
	v1service "gin/user-management-api/internal/service/v1"
	"gin/user-management-api/pkg/auth"
	"gin/user-management-api/pkg/cache"
	"gin/user-management-api/pkg/mail"
)

type AuthModule struct {
	routes routes.Route
}

func NewAuthModule(ctx *MouldeContext, tokenService auth.TokenService, cacheService cache.RedisCacheService, mailService mail.EmailProviderService) *AuthModule {
	// Initialize the auth repository
	userRepository := repository.NewSqlUserRepository(ctx.DB)

	// Initialize the auth services
	authService := v1service.NewAuthService(userRepository, tokenService, cacheService, mailService)

	// Initialize the auth handler
	authHandler := v1handler.NewAuthHandler(authService)

	// Initialize the auth routes
	authRoutes := v1routes.NewAuthRoutes(authHandler)

	return &AuthModule{routes: authRoutes}
}

func (m *AuthModule) Routes() routes.Route {
	return m.routes
}
