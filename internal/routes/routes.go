package routes

import (
	"gin/user-management-api/internal/middleware"
	v1routes "gin/user-management-api/internal/routes/v1"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/auth"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)


type Route interface {
	Register(r *gin.RouterGroup)
}


func RegisterRoutes(r *gin.Engine, authService auth.TokenService, routes ...Route) {
	httpLogger := utils.NewLoggerWithPath("../../internal/logs/http.log", "info")
	recoveryLogger := utils.NewLoggerWithPath("../../internal/logs/recovery.log", "warning")
	rateLimiterLogger := utils.NewLoggerWithPath("../../internal/logs/rate_limiter.log", "warning")

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(
				middleware.RateLimiterMiddleware(rateLimiterLogger),
				middleware.TraceMiddleware(),
				middleware.LoggerMiddleware(httpLogger),
				middleware.RecoveryMiddleware(recoveryLogger),
				middleware.ApiKeyMiddleware(),
			)
	v1api := r.Group("/api/v1")

	middleware.InitAuthMiddleware(authService)

	protected := v1api.Group("")
	protected.Use(
		middleware.AuthMiddleware(),
	)

	for _, route := range routes {
		switch route.(type) {
		case *v1routes.AuthRoutes:
			route.Register(v1api)
		default:
			route.Register(protected)
		}
	}

	r.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(404, gin.H{
			"error": "Not found",
			"path": ctx.Request.URL.Path,
		})
	})
}
