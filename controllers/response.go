package controllers

import "github.com/gin-gonic/gin"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

type APIResponse struct {
	OK      bool        `json:"ok"`
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

func respondSuccess(ctx *gin.Context, status int, message string, data interface{}) {
	ctx.JSON(status, APIResponse{
		OK:      true,
		Status:  status,
		Message: message,
		Data:    data,
	})
}

func respondError(ctx *gin.Context, status int, code string, message string, detail string) {
	ctx.JSON(status, APIResponse{
		OK:      false,
		Status:  status,
		Message: message,
		Error: &APIError{
			Code:    code,
			Message: message,
			Detail:  detail,
		},
	})
}
