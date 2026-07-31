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
	noteAccessHandler *NoteAccessHandler,
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

	// --- Note Access (Owner Only): kasih/ubah/cabut/lihat akses viewer
	// atau editor ke user lain lewat email. AccessMiddleware dipasang di
	// sini BUKAN untuk membatasi Owner-only-nya (itu sudah dicek ulang
	// sendiri di NoteAccessUsecase, independen dari middleware), tapi
	// supaya note-nya ke-resolve dan tersedia lewat c.Locals("note").
	notes.Post("/:id/access",
		middleware.AuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteAccessHandler.Grant,
	)
	notes.Delete("/:id/access",
		middleware.AuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteAccessHandler.Revoke,
	)
	notes.Get("/:id/access",
		middleware.AuthMiddleware(jwtSecret),
		middleware.AccessMiddleware(noteRepo, noteAccessRepo),
		noteAccessHandler.List,
	)
}