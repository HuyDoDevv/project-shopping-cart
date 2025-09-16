package v1handler

import (
	v1dto "gin/user-management-api/internal/dto/v1"
	v1service "gin/user-management-api/internal/service/v1"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)


type UserHandler struct {
	service v1service.UserService
}

func NewUserHandler(service v1service.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (uh *UserHandler) GetAllUsers(ctx *gin.Context) {
	var params v1dto.GetUsersParamSearch
	if err := ctx.ShouldBindQuery(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	users, countUser, err := uh.service.GetAllUsers(ctx, params.Search, params.Order, params.Sort, params.Page, params.Limit, false)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	usersDto := v1dto.MapUsersToDTO(users)
	paginationResp := utils.NewPaginationResponse(usersDto, params.Page, params.Limit, countUser)
	utils.ResponseSuccess(ctx, http.StatusOK, "Get all user successfully", paginationResp)
}

func (uh *UserHandler) CreateUser(ctx *gin.Context) {
	var input v1dto.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}
	user := input.MapCreaeteInputToModel()
	createdUser , err := uh.service.CreateUser(ctx, user)

	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	userDto := v1dto.MapUserToDTO(createdUser)

	utils.ResponseSuccess(ctx, http.StatusCreated, "User created successfully" ,userDto)
}

func (uh *UserHandler) GetUserByUUID(ctx *gin.Context) {
	var params v1dto.GetUserByUuidParams
	if err := ctx.ShouldBindUri(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	uuidUser, err := uuid.Parse(params.Uuid);
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	user, err := uh.service.GetUserByUUID(ctx, uuidUser)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	userDto := v1dto.MapUserToDTO(user)
	utils.ResponseSuccess(ctx, http.StatusOK, "find user successfully",userDto)
}

func (uh *UserHandler) UpdateUser(ctx *gin.Context) {
	var params v1dto.GetUserByUuidParams
	if err := ctx.ShouldBindUri(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	uuidUser, err := uuid.Parse(params.Uuid);
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}


	var input v1dto.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	user := input.MapUpdateInputToModel(uuidUser)
	updateUser, err := uh.service.UpdateUser(ctx, user)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	userDto := v1dto.MapUserToDTO(updateUser)
	utils.ResponseSuccess(ctx, http.StatusCreated, "User updated successfully",userDto)
}

func (uh *UserHandler) SortDeleteUser(ctx *gin.Context) {
	var params v1dto.GetUserByUuidParams
	if err := ctx.ShouldBindUri(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	uuidUser, err := uuid.Parse(params.Uuid);
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	user, err := uh.service.SoftDeleteUser(ctx, uuidUser)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	userDto := v1dto.MapUserToDTO(user)
	utils.ResponseSuccess(ctx, http.StatusOK, "User deleted successfully",userDto)
}

func (uh *UserHandler) RestoreUser(ctx *gin.Context) {
	var params v1dto.GetUserByUuidParams
	if err := ctx.ShouldBindUri(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	uuidUser, err := uuid.Parse(params.Uuid);
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	user, err := uh.service.RestoreUser(ctx, uuidUser)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	userDto := v1dto.MapUserToDTO(user)
	utils.ResponseSuccess(ctx, http.StatusOK, "Resote user successfully",userDto)
}

func (uh *UserHandler) DeleteUser(ctx *gin.Context) {
	var params v1dto.GetUserByUuidParams
	if err := ctx.ShouldBindUri(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	uuidUser, err := uuid.Parse(params.Uuid);
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	err = uh.service.DeleteUser(ctx, uuidUser)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}

	utils.ResponseStatusCode(ctx, http.StatusOK)
}

func (uh *UserHandler) GetUserSoftDeleted(ctx *gin.Context) {
	var params v1dto.GetUsersParamSearch
	if err := ctx.ShouldBindQuery(&params); err != nil {
		utils.ResponseValidation(ctx, validation.HandleValidationErrors(err))
		return
	}

	users, countUser, err := uh.service.GetAllUsers(ctx, params.Search, params.Order, params.Sort, params.Page, params.Limit, true)
	if err != nil {
		utils.ResponseError(ctx, err)
		return
	}
	usersDto := v1dto.MapUsersToDTO(users)
	paginationResp := utils.NewPaginationResponse(usersDto, params.Page, params.Limit, countUser)
	utils.ResponseSuccess(ctx, http.StatusOK, "Get all user soft deleted successfully", paginationResp)
}
