package middleware

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/token"
	"github.com/gofiber/fiber/v2"
)

// OptionalAuthMiddleware itu SAUDARA dari AuthMiddleware, bedanya cuma
// satu: dia TIDAK PERNAH menolak request. Kalau cookie token ADA dan
// valid, user_id diselipkan ke c.Locals (sama seperti AuthMiddleware).
// Kalau token TIDAK ADA atau tidak valid, request tetap lanjut (c.Next())
// dengan c.Locals("user_id") kosong — dianggap sebagai "guest".
//
// Dipakai khusus untuk endpoint yang boleh diakses TANPA login (mis.
// GET /notes/:slug untuk note public), tapi tetap perlu tahu "siapa yang
// akses" KALAU KEBETULAN dia sedang login (supaya AccessMiddleware bisa
// mendeteksi "oh ini kan Owner-nya sendiri").
func OptionalAuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("token")
		if tokenString == "" {
			return c.Next() // tidak ada cookie -> lanjut sebagai guest, BUKAN error
		}

		claims, err := token.Parse(tokenString, jwtSecret)
		if err != nil {
			return c.Next() // token ada tapi rusak/kadaluarsa -> tetap lanjut sebagai guest
		}

		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
