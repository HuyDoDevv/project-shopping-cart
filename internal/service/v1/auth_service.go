package v1service

import (
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/auth"
	"gin/user-management-api/pkg/cache"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

type authService struct {
	userRepo     repository.UserRepository
	tokenService auth.TokenService
	cache        cache.RedisCacheService
}

type LoginAttempt struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu              sync.Mutex
	clients         = make(map[string]*LoginAttempt)
	LoginAttemptTTL = 5 * time.Minute
	MaxLoginAttempt = 5
)

func NewAuthService(repo repository.UserRepository, tokenService auth.TokenService, cache cache.RedisCacheService) *authService {
	return &authService{
		userRepo:     repo,
		tokenService: tokenService,
		cache:        cache,
	}
}

func (as *authService) getClientIP(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}

	return ip
}

func (as *authService) getLoginAttempt(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	client, exists := clients[ip]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(float32(MaxLoginAttempt)/float32(LoginAttemptTTL.Seconds())), MaxLoginAttempt) // 5 request/sec, brust 10
		newClient := &LoginAttempt{limiter, time.Now()}
		clients[ip] = newClient
		return limiter
	}

	client.lastSeen = time.Now()
	return client.limiter
}

func (as *authService) checkLoginAttempt(ip string) error {
	limiter := as.getLoginAttempt(ip)
	if !limiter.Allow() {
		return utils.NewError(utils.TooManyRequestsError, "Too many login attempt. Please rety again later")
	}
	return nil
}

func (as *authService) CleanupClients(ip string) {
	mu.Lock()
	defer mu.Unlock()
	delete(clients, ip)
}

func (as *authService) Login(ctx *gin.Context, email, password string) (string, string, int, error) {
	context := ctx.Request.Context()
	ip := as.getClientIP(ctx)

	if err := as.checkLoginAttempt(ip); err != nil {
		return "", "", 0, err
	}

	email = utils.NormalizeString(email)
	user, err := as.userRepo.GetByEmail(context, email)
	if err != nil {
		as.getLoginAttempt(ip)
		return "", "", 0, utils.NewError(utils.UnauthorizedError, "Invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword), []byte(password)); err != nil {
		return "", "", 0, utils.NewError(utils.UnauthorizedError, "Invalid email or password")
	}

	accessToken, err := as.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "unable to create access token", err)
	}

	refreshTokenToken, err := as.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "unable to create access token", err)
	}

	if err := as.tokenService.StoreRefreshToken(refreshTokenToken); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "Cannot save refresh token", err)
	}

	as.CleanupClients(ip)

	return accessToken, refreshTokenToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}

func (as *authService) RefreshToken(ctx *gin.Context, refreshTokenString string) (string, string, int, error) {
	context := ctx.Request.Context()

	// kiểm tra refresh token, để trả về uuid của user
	token, err := as.tokenService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", 0, utils.NewError(utils.UnauthorizedError, "Refresh token is invalid or revoked")
	}

	// Lấy thông tin uuid của user
	userUuid, _ := uuid.Parse(token.UserUUID)
	user, err := as.userRepo.FindByUUID(context, userUuid)
	if err != nil {
		return "", "", 0, utils.NewError(utils.UnauthorizedError, "User not found")
	}

	// Tạo access token mới
	accessToken, err := as.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "unable to create access token", err)
	}

	// Tạo refresh token mới

	refreshTokenToken, err := as.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "unable to create access token", err)
	}

	// Vô hiệu hoá refresh token cũ
	if err := as.tokenService.RevokeRefreshToken(refreshTokenString); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "Unable to revoke token", err)
	}

	// Lưu refeshtoken mới
	if err := as.tokenService.StoreRefreshToken(refreshTokenToken); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError, "Cannot save refresh token", err)
	}

	return accessToken, refreshTokenToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}

func (as *authService) Logout(ctx *gin.Context, refreshToken string) error {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return utils.NewError(utils.UnauthorizedError, "Missing Authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, claims, err := as.tokenService.ParseToken(tokenString)
	if err != nil {
		return utils.NewError(utils.UnauthorizedError, "Invalid access token")
	}

	if jti, ok := claims["jti"].(string); ok {
		expUnit, _ := claims["exp"].(float64)
		exp := time.Unix(int64(expUnit), 0)
		key := "backlist:" + jti
		ttl := time.Until(exp)
		as.cache.Set(key, "revoked", ttl)
	}

	_, err = as.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return utils.NewError(utils.UnauthorizedError, "Refresh token is invalid or revoked")
	}

	if err := as.tokenService.RevokeRefreshToken(refreshToken); err != nil {
		return utils.WrapError(utils.InternalServerError, "Invalid access token", err)
	}

	return nil
}
