package controllers

import (
	"cobago/services"
	"cobago/utils"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	authService services.AuthService
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Format request tidak sesuai", err.Error())
	}

	if req.Email == "" || req.Password == "" {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Email dan password wajib diisi", nil)
	}

	accessToken, refreshToken, err := c.authService.Login(ctx.Context(), req.Email, req.Password)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusUnauthorized, err.Error(), nil)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	type RefreshRequest struct {
		Refresh_Token string `json:"refresh_token"`
	}

	var req RefreshRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.SendError(ctx, fiber.StatusBadRequest, "Format request tidak sesuai", err.Error())
	}

	newAccessToken, err := c.authService.Refresh(ctx.Context(), req.Refresh_Token)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusUnauthorized, err.Error(), nil)
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Token berhasil diperbarui", fiber.Map{
		"access_token": newAccessToken,
	})
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(uint)
	if !ok {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Sesi tidak valid", nil)
	}

	if err := c.authService.Logout(ctx.Context(), userID); err != nil {
		return utils.SendError(ctx, fiber.StatusInternalServerError, "Gagal melakukan logout", err.Error())
	}

	return utils.SendSuccess(ctx, fiber.StatusOK, "Logout berhasil, sesi telah dihapus", nil)
}