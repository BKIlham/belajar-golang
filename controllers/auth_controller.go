package controllers

import (
	"cobago/services"
	"cobago/utils"
	"time"

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

	// SUNTIKKAN REFRESH TOKEN KE HTTPONLY COOKIE (ANTI-XSS & ANTI-CSRF VIA SAMESITE)
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,                          // JavaScript FE tidak bisa baca
		Secure:   false,                         // Set ke true jika di production (HTTPS), lokal cukup false
		SameSite: fiber.CookieSameSiteLaxMode,   // Menghadang serangan CSRF dari web lain
		Path:     "/",                           // Berlaku di semua path aplikasi
	})

	// Kembalikan HANYA Access Token ke JSON Body
	return utils.SendSuccess(ctx, fiber.StatusOK, "Login berhasil", fiber.Map{
		"access_token": accessToken,
	})
}

func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	// 🛡️ BACA REFRESH TOKEN LANGSUNG DARI COOKIE BROWSER, BUKAN DARI JSON BODY LAGI
	refreshToken := ctx.Cookies("refresh_token")
	if refreshToken == "" {
		return utils.SendError(ctx, fiber.StatusUnauthorized, "Sesi berakhir, silakan login kembali", nil)
	}

	// Jalankan service (sekarang merotasi token dan mengembalikan sepasang token baru)
	newAccessToken, newRefreshToken, err := c.authService.Refresh(ctx.Context(), refreshToken)
	if err != nil {
		return utils.SendError(ctx, fiber.StatusUnauthorized, err.Error(), nil)
	}

	//  SET ULANG COOKIE DENGAN REFRESH TOKEN YANG BARU (ROTATION)
	ctx.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   false,
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/",
	})

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

	// HAPUS COOKIE REFRESH TOKEN SAAT LOGOUT
	ctx.ClearCookie("refresh_token")

	return utils.SendSuccess(ctx, fiber.StatusOK, "Logout berhasil, sesi telah dihapus", nil)
}
