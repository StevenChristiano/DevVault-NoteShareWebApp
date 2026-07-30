package http

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/delivery/http/middleware"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendaftarkan semua route API. Fungsi ini yang jadi "peta"
// URL -> handler mana yang menangani. Sengaja dipisah dari main.go supaya
// main.go tetap tipis (cuma wiring), dan penambahan route baru di
// Tahap 3+ dilakukan di sini, bukan di main.go.
func SetupRoutes(
	app *fiber.App, 
	authHandler *AuthHandler,
	noteHandler *NoteHandler,
	noteRepo repository.NoteRepository,
	noteAccessRepo repository.NoteAccessRepository,
	jwtSecret string,
) {
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Note routes
	notes := api.Group("/notes")
	notes.Post("/", middleware.AuthMiddleware(jwtSecret), noteHandler.Create)
	notes.Get("/", middleware.AuthMiddleware(jwtSecret), noteHandler.ListMine)

	notes.Put("/:id",
		middleware.AuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteHandler.Update,
	)
	notes.Delete("/:id",
		middleware.AuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteHandler.Delete,
	)
 
	notes.Get("/:slug",
		middleware.OptionalAuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteHandler.GetBySlug,
	)

	//another routes
}