package middleware

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/token"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware menjawab pertanyaan "SIAPA kamu?" — bukan "kamu boleh
// akses apa?" (itu tugas AccessMiddleware, lihat file sebelahnya).
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("token")
		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not logged in"})
		}

		claims, err := token.Parse(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals("user_id", claims.UserID) //save user_id to locals so that next handler can use it (not permanent, only for this request)

		return c.Next()
	}
}
