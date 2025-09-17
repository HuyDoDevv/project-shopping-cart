package v1service

import (
	"gin/user-management-api/internal/db/sqlc"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


type UserService interface {
	GetAllUsers(ctx *gin.Context, search string, orderBy, sort string, page, limit int32, deleted bool) ([]sqlc.User, int32,error)
	CreateUser(ctx *gin.Context, user sqlc.CreateUserParams) (sqlc.User, error)
	GetUserByUUID(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error)
	UpdateUser(ctx *gin.Context, userParams sqlc.UpdateUserByUuidParams) (sqlc.User, error)
	SoftDeleteUser(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error)
	RestoreUser(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error)
	DeleteUser(ctx *gin.Context, userUuid uuid.UUID) error
}


type AuthService interface {
	Login(ctx *gin.Context, email, password string) (string, string, int, error)
	Logout(ctx *gin.Context, refreshToken string) error
	RefreshToken(ctx *gin.Context, token string) (string, string, int, error)
}
