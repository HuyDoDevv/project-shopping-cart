package v1service

import (
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)


type authService struct {
	userRepo repository.UserRepository
	tokenService auth.TokenService
}


func NewAuthService(repo repository.UserRepository, tokenService auth.TokenService) *authService {
	return &authService{
		userRepo: repo,
		tokenService: tokenService,
	}
}


func (as *authService) Login(ctx *gin.Context, email, password string) (string, string, int, error) {
	context := ctx.Request.Context()
	email = utils.NormalizeString(email)
	user , err := as.userRepo.GetByEmail(context,email)
	if err != nil {
		return "", "", 0, utils.NewError(utils.UnauthorizedError,"Invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword),[]byte(password)); err != nil {
		return "", "", 0, utils.NewError(utils.UnauthorizedError,"Invalid email or password")
	}

	accessToken, err := as.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"unable to create access token", err)
	}

	refreshTokenToken, err := as.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"unable to create access token", err)
	}

	if err := as.tokenService.StoreRefreshToken(refreshTokenToken); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"Cannot save refresh token", err)
	}

	return accessToken, refreshTokenToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}

func (as *authService) Logout(ctx *gin.Context) error {
	return nil
}

func (as *authService) RefreshToken(ctx *gin.Context, refreshTokenString string) (string, string, int, error) {
	context := ctx.Request.Context()

	// kiểm tra refresh token, để trả về uuid của user
	token, err := as.tokenService.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "" ,"", 0, utils.NewError(utils.UnauthorizedError,"Refresh token is invalid or revoked")
	}

	// Lấy thông tin uuid của user
	userUuid, _ := uuid.Parse(token.UserUUID)
	user, err := as.userRepo.FindByUUID(context,userUuid)
	if err != nil {
		return "" ,"", 0, utils.NewError(utils.UnauthorizedError,"User not found")
	}

	// Tạo access token mới
	accessToken, err := as.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"unable to create access token", err)
	}

	// Tạo refresh token mới

	refreshTokenToken, err := as.tokenService.GenerateRefreshToken(user)
	if err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"unable to create access token", err)
	}

	// Vô hiệu hoá refresh token cũ
	if err := as.tokenService.RevokeRefreshToken(refreshTokenString); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"Unable to revoke token", err)
	}

	// Lưu refeshtoken mới
	if err := as.tokenService.StoreRefreshToken(refreshTokenToken); err != nil {
		return "", "", 0, utils.WrapError(utils.InternalServerError,"Cannot save refresh token", err)
	}

	return accessToken, refreshTokenToken.Token, int(auth.AccessTokenTTL.Seconds()), nil
}
