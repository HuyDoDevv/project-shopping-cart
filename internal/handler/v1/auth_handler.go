package v1handler

import (
	v1dto "gin/user-management-api/internal/dto/v1"
	v1service "gin/user-management-api/internal/service/v1"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service v1service.AuthService
}

func NewAuthHandler(service v1service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}


func (ah *AuthHandler) Login(ctx *gin.Context){
	var input v1dto.LoginInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	accessToken, refreshToken, expiresIn, err := ah.service.Login(ctx, input.Email, input.Password)

	if err != nil {
		utils.ResponseError(ctx,err)
		return
	}

	response := v1dto.LoginResponse {
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		ExpiresIn: expiresIn,
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "Login successfully", response)
}

func (ah *AuthHandler) Logout(ctx *gin.Context){
	var input v1dto.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "Logout successfully")
}


func (ah *AuthHandler) RefreshToken(ctx *gin.Context){
	var input v1dto.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	accessToken, refreshToken, expiresIn, err := ah.service.RefreshToken(ctx, input.RefreshToken)

	if err != nil {
		utils.ResponseError(ctx,err)
		return
	}

	response := v1dto.LoginResponse {
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		ExpiresIn: expiresIn,
	}

	utils.ResponseSuccess(ctx, http.StatusOK, "Refresh Token successfully", response)
}
