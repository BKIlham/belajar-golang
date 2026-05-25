package utils

import "github.com/gofiber/fiber/v2"


type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` 
}


func SendSuccess(ctx *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return ctx.Status(statusCode).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}


func SendError(ctx *fiber.Ctx, statusCode int, message string, detailError interface{}) error {
	return ctx.Status(statusCode).JSON(WebResponse{
		Success: false,
		Message: message,
		Data:    detailError, 
	})
}