package v1dto

import (
	"gin/user-management-api/internal/db/sqlc"
	"gin/user-management-api/internal/utils"

	"github.com/google/uuid"
)

type UserDTO struct {
	UUID      string `json:"uuid"`
	Name			string `json:"full_name"`
	Email     string `json:"email_address"`
	Age			 	*int    `json:"age"`
	Status    string `json:"status"`
	Level     string `json:"level" `
	CreatedAt string `json:"craeted_at" `
}

type CreateUserInput struct {
	Name 						string 			`json:"name" binding:"required"`
	Email 					string 			`json:"email" binding:"required,email,email_advanced"`
	Age 						int32				`json:"age" binding:"gt=0"`
	Password 				string 			`json:"password" binding:"required,min=8,password_strong"`
	Status 					int32 				`json:"status" binding:"required,oneof=1 2 3"`
	Level 					int32 				`json:"level" binding:"required,oneof=1 2 3"`
}


type GetUsersParamSearch struct {
	Search 	string 			`form:"search" binding:"omitempty,min=3,max=50,search"`
	Page 		int32 			`form:"page" binding:"omitempty,gte=1"`
	Limit 	int32 			`form:"limit" binding:"omitempty,gte=1,lte=500"`
	Order 	string 			`form:"order_by" binding:"omitempty,oneof=user_id user_created_at"`
	Sort 		string 			`form:"sort" binding:"omitempty,oneof=asc desc"`
}

type GetUserByUuidParams struct {
	Uuid string `uri:"uuid" binding:"uuid"`
}

func (input *CreateUserInput) MapCreaeteInputToModel() sqlc.CreateUserParams {
	return sqlc.CreateUserParams{
		UserEmail: input.Email,
		UserFullname: input.Name,
		UserPassword: input.Password,
		UserAge: utils.ConvertToInt32Pointer(input.Age),
		UserStatus: input.Status,
		UserLevel: input.Level,
	}
}

type UpdateUserInput struct {
	Name 						*string 			`json:"name" binding:"omitempty"`
	Age 						*int32				`json:"age" binding:"omitempty,gt=0"`
	Password 				*string 			`json:"password" binding:"omitempty,min=8,password_strong"`
	Status 					*int32 				`json:"status" binding:"omitempty,oneof=1 2 3"`
	Level 					*int32 				`json:"level" binding:"omitempty,oneof=1 2 3"`
}

func (input *UpdateUserInput) MapUpdateInputToModel(userUuid uuid.UUID) sqlc.UpdateUserByUuidParams {
	return sqlc.UpdateUserByUuidParams{
		UserFullname: input.Name,
		UserPassword: input.Password,
		UserAge: input.Age,
		UserStatus: input.Status,
		UserLevel: input.Level,
		UserUuid: userUuid,
	}
}

func MapUserToDTO(user sqlc.User) *UserDTO {
	dto := &UserDTO{
		UUID: user.UserUuid.String(),
		Name: user.UserFullname,
		Email: user.UserEmail,
		Status: mapStatusToString(int(user.UserStatus)),
		Level: mapLevelToString(int(user.UserLevel)),
		CreatedAt: user.UserCreatedAt.Format("2006-01-01 15:01:01"),
	}
	if user.UserAge != nil {
		age := int(*user.UserAge)
		dto.Age = & age
	}
	return dto
}

func MapUsersToDTO(users []sqlc.User) []UserDTO {
	dtos := make([]UserDTO, 0,len(users))
	for _,user := range users{
		dtos = append(dtos, *MapUserToDTO(user))
	}
	return dtos
}

func mapStatusToString(status int) string {
	switch status {
	case 1:
		return "Active"
	case 2:
		return "Inactive"
	case 3:
		return "Banned"
	default:
		return "None"
	}
}

func mapLevelToString(level int) string {
	switch level {
	case 1:
		return "Administrator"
	case 2:
		return "Moderator"
	case 3:
		return "Member"
	default:
		return "None"
	}
}
