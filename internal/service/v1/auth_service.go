package v1service

import (
	"gin/user-management-api/internal/repository"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/auth"

	"github.com/gin-gonic/gin"
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


func (as *authService) Login(ctx *gin.Context, email, password string) (string, int, error) {
	context := ctx.Request.Context()
	email = utils.NormalizeString(email)
	user , err := as.userRepo.GetByEmail(context,email)
	if err != nil {
		return "", 0, utils.NewError(utils.UnauthorizedError,"Invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.UserPassword),[]byte(password)); err != nil {
		return "", 0, utils.NewError(utils.UnauthorizedError,"Invalid email or password")
	}

	accessToken, err := as.tokenService.GenerateAccessToken(user)
	if err != nil {
		return "", 0, utils.WrapError(utils.InternalServerError,"unable to create access token", err)
	}

	return accessToken, int(auth.AccessTokenTTL.Seconds()), nil
}

func (as *authService) Logout(ctx *gin.Context) error {
	return nil
}
