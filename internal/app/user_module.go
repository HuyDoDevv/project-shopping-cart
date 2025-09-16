package app

import (
	v1handler "gin/user-management-api/internal/handler/v1"
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/routes"
	v1routes "gin/user-management-api/internal/routes/v1"
	v1service "gin/user-management-api/internal/service/v1"
)

type UserModule struct {
	routes routes.Route

}

func NewUserModule(ctx *MouldeContext) *UserModule {
	// Initialize the user repository
	userRepository := repository.NewSqlUserRepository(ctx.DB)

	// Initialize the user services
	userService := v1service.NewUserService(userRepository, ctx.Redis)

	// Initialize the user handler
	userHandler := v1handler.NewUserHandler(userService)

	// Initialize the user routes
	userRoutes := v1routes.NewUserRoutes(userHandler)

	return &UserModule{routes: userRoutes}
}

func (m *UserModule) Routes() routes.Route {
	return m.routes
}
