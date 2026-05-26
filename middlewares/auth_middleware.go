package middlewares

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah middleware untuk memvalidasi Access Token JWT
func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Missing authorization token",
			})
		}

		// 2. Format harus "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid authorization format. Use 'Bearer <token>'",
			})
		}

		tokenStr := parts[1]

		// 3. Parse dan validasi token menggunakan SECRET_KEY
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			// Pastikan algoritma signing-nya sesuai
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte(os.Getenv("JWT_ACCESS_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Invalid or expired access token",
			})
		}

		// 4. Ekstrak claims data user dari token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Failed to parse token claims",
			})
		}

		// 5. Simpan user_id ke context lokal Fiber agar bisa dibaca di layer Controller
		userID := uint(claims["user_id"].(float64))
		c.Locals("user_id", userID)

		// Lolos validasi, lanjutkan ke handler/controller berikutnya
		return c.Next()
	}
}