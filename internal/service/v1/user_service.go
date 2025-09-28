package v1service

import (
	"database/sql"
	"errors"
	"fmt"
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/cache"
	"gin/user-management-api/pkg/loggers"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repository repository.UserRepository
	cache      cache.RedisCacheService
}

func NewUserService(repository repository.UserRepository, redisClient *redis.Client) UserService {
	return &userService{
		repository: repository,
		cache:      cache.NewRedisCacheService(redisClient),
	}
}

func (us *userService) GetAllUsers(ctx *gin.Context, search, orderBy, sort string, page, limit int32, deleted bool) ([]sqlc.User, int32, error) {
	context := ctx.Request.Context()

	if sort == "" {
		sort = "desc"
	}

	if orderBy == "" {
		orderBy = "user_created_at"
	}

	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		envLimit := utils.GetIntEnv("LIMIT_ITEM_ON_PER_PAGE", 10)
		limit = int32(envLimit)
	}
	offset := (page - 1) * limit

	//Get cache data
	var cacheKey = us.generateCacheKey(search, orderBy, sort, page, limit, deleted)

	var cacheData struct {
		User  []sqlc.User `json:"users"`
		Total int32       `json:"total"`
	}

	if err := us.cache.Get(cacheKey, &cacheData); err == nil && cacheData.User != nil {
		return cacheData.User, cacheData.Total, nil
	}

	users, err := us.repository.GetAllV2(context, search, orderBy, sort, limit, offset, deleted)
	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(utils.InternalServerError, "failed to get all user", err)
	}

	total, err := us.repository.CountUsers(context, search, deleted)
	if err != nil {
		return []sqlc.User{}, 0, utils.WrapError(utils.InternalServerError, "failed to count user", err)
	}

	// Create cache data
	cacheData = struct {
		User  []sqlc.User `json:"users"`
		Total int32       `json:"total"`
	}{
		User:  users,
		Total: int32(total),
	}

	us.cache.Set(cacheKey, cacheData, 5*time.Minute)

	return users, int32(total), nil
}

func (us *userService) CreateUser(ctx *gin.Context, intUserParams sqlc.CreateUserParams) (sqlc.User, error) {
	context := ctx.Request.Context()

	intUserParams.UserEmail = utils.NormalizeString(intUserParams.UserEmail)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(intUserParams.UserPassword), bcrypt.DefaultCost)
	if err != nil {
		return sqlc.User{}, utils.WrapError(utils.InternalServerError, "failed to hash password", err)
	}

	intUserParams.UserPassword = string(hashedPassword)

	user, err := us.repository.Create(context, intUserParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, utils.NewError(utils.ConflictError, "Email already exitst")
		}

		return sqlc.User{}, utils.WrapError(utils.InternalServerError, "failed to create a new user", err)
	}

	if err := us.cache.Clear("users:*"); err != nil {
		loggers.Log.Warn().Err(err).Msg("Failed to clear cache")
	}

	return user, nil
}

func (us *userService) GetUserByUUID(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context()
	user, err := us.repository.FindByUUID(context, userUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.WrapError(utils.NotFoundError, "user not found", err)
		}
		return sqlc.User{}, utils.NewError(utils.ConflictError, "failed to get an user")
	}
	return user, nil
}

func (us *userService) UpdateUser(ctx *gin.Context, userParams sqlc.UpdateUserByUuidParams) (sqlc.User, error) {
	context := ctx.Request.Context()

	if userParams.UserPassword != nil && *userParams.UserPassword != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*userParams.UserPassword), bcrypt.DefaultCost)
		if err != nil {
			return sqlc.User{}, utils.WrapError(utils.InternalServerError, "failed to hash password", err)
		}
		hashed := string(hashedPassword)
		userParams.UserPassword = &hashed
	}

	userUpdate, err := us.repository.Update(context, userParams)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.WrapError(utils.NotFoundError, "user not found", err)
		}
		return sqlc.User{}, utils.WrapError(utils.InternalServerError, "failed to update user", err)
	}

	if err := us.cache.Clear("users:*"); err != nil {
		loggers.Log.Warn().Err(err).Msg("Failed to clear cache")
	}

	return userUpdate, nil
}

func (us *userService) SoftDeleteUser(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context()
	user, err := us.repository.SoftDelete(context, userUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.WrapError(utils.NotFoundError, "user not found", err)
		}
		return sqlc.User{}, utils.NewError(utils.ConflictError, "failed to delete user")
	}
	if err := us.cache.Clear("users:*"); err != nil {
		loggers.Log.Warn().Err(err).Msg("Failed to clear cache")
	}
	return user, nil
}

func (us *userService) RestoreUser(ctx *gin.Context, userUuid uuid.UUID) (sqlc.User, error) {
	context := ctx.Request.Context()
	user, err := us.repository.Restore(context, userUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.User{}, utils.WrapError(utils.NotFoundError, "user not found", err)
		}
		return sqlc.User{}, utils.NewError(utils.ConflictError, "failed to restore user")
	}
	if err := us.cache.Clear("users:*"); err != nil {
		loggers.Log.Warn().Err(err).Msg("Failed to clear cache")
	}
	return user, nil
}

func (us *userService) DeleteUser(ctx *gin.Context, userUuid uuid.UUID) error {
	context := ctx.Request.Context()
	_, err := us.repository.Delete(context, userUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.WrapError(utils.NotFoundError, "user not found", err)
		}
		return utils.NewError(utils.ConflictError, "failed to delete user")
	}
	if err := us.cache.Clear("users:*"); err != nil {
		loggers.Log.Warn().Err(err).Msg("Failed to clear cache")
	}
	return nil
}

func (us *userService) generateCacheKey(search, orderBy, sort string, page, limit int32, deleted bool) string {
	search = strings.TrimSpace(search)
	if search == "" {
		search = "none"
	}

	orderBy = strings.TrimSpace(orderBy)
	if orderBy == "" {
		orderBy = "user_created_at"
	}

	sort = strings.ToLower(strings.TrimSpace(sort))
	if sort == "" {
		sort = "desc"
	}
	return fmt.Sprintf("users:%s:%s:%s:%d:%d:%t", search, orderBy, sort, page, limit, deleted)
}
