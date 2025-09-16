package v1routes

import (
	v1handler "gin/user-management-api/internal/handler/v1"

	"github.com/gin-gonic/gin"
)


type UserRoutes struct {
	handler *v1handler.UserHandler
}

func NewUserRoutes(handler *v1handler.UserHandler) *UserRoutes {
	return &UserRoutes{
		handler: handler,
	}
}

func (ur *UserRoutes) Register(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.GET("/", ur.handler.GetAllUsers)
		users.GET("/soft-deleted", ur.handler.GetUserSoftDeleted)
		users.POST("/", ur.handler.CreateUser)
		users.GET("/:uuid", ur.handler.GetUserByUUID)
		users.PUT("/:uuid", ur.handler.UpdateUser)
		users.DELETE("/:uuid", ur.handler.SortDeleteUser)
		users.PATCH("/:uuid/restore", ur.handler.RestoreUser)
		users.DELETE("/:uuid/trash", ur.handler.DeleteUser)
	}
}
