package http

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan semua route API. Fungsi ini yang jadi "peta"
// URL -> handler mana yang menangani. Sengaja dipisah dari main.go supaya
// main.go tetap tipis (cuma wiring), dan penambahan route baru di
// Tahap 3+ dilakukan di sini, bukan di main.go.
func SetupRoutes(app *fiber.App, authHandler *AuthHandler) {
	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Endpoint notes (CRUD, like, save, dst) menyusul di Tahap 3, memakai
	// middleware.AuthMiddleware untuk endpoint yang wajib login dan
	// middleware.AccessMiddleware untuk endpoint GET /notes/:slug yang
	// aksesnya tergantung visibility.
}