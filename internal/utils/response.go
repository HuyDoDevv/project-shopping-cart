package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorCode string

const (
	NotFoundError        ErrorCode = "NOT_FOUND"
	InternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
	BadRequestError      ErrorCode = "BAD_REQUEST"
	UnauthorizedError    ErrorCode = "UNAUTHORIZED"
	ForbiddenError       ErrorCode = "FORBIDDEN"
	ConflictError        ErrorCode = "CONFLICT"
	TooManyRequestsError ErrorCode = "TOO_MANY_REQUESTS"
)

type AppError struct {
	Message string
	Code    ErrorCode
	Err     error
}

type APIResponse struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Data       any    `json:"data,omitempty"`
	Pagination any    `json:"pagination,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(code ErrorCode, message string) error {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func WrapError(code ErrorCode, message string, err error) error {
	if err == nil {
		return nil
	}
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func ResponseError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		statusCode := httpStatusFromErrorCode(appErr.Code)
		response := gin.H{"error": CapitalizrFirst(appErr.Message), "code": appErr.Code}
		if appErr.Err != nil {
			response["details"] = appErr.Err.Error()
		}
		ctx.JSON(statusCode, response)
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  InternalServerError,
	})
}

func ResponseSuccess(ctx *gin.Context, status int, message string, data ...any) {
	resp := APIResponse{
		Status:  "success",
		Message: CapitalizrFirst(message),
	}

	if len(data) > 0 && data[0] != nil {
		if m, ok := data[0].(map[string]any); ok {
			if p, exitsts := m["pagination"]; exitsts {
				resp.Pagination = p
			}

			if d, exitsts := m["data"]; exitsts {
				resp.Data = d
			} else {
				resp.Data = m
			}
		} else {
			resp.Data = data[0]
		}
	}
	ctx.JSON(status, resp)
}

func ResponseStatusCode(ctx *gin.Context, status int) {
	ctx.Status(status)
}

func ResponseValidation(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusBadRequest, data)
}

func httpStatusFromErrorCode(code ErrorCode) int {
	switch code {
	case NotFoundError:
		return http.StatusNotFound
	case InternalServerError:
		return http.StatusInternalServerError
	case BadRequestError:
		return http.StatusBadRequest
	case UnauthorizedError:
		return http.StatusUnauthorized
	case ForbiddenError:
		return http.StatusForbidden
	case ConflictError:
		return http.StatusConflict
	case TooManyRequestsError:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
