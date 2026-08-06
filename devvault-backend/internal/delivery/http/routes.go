package http

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/delivery/http/middleware"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// RouteDeps membungkus SEMUA dependency yang dibutuhkan SetupRoutes,
// menggantikan daftar parameter panjang satu-satu. Kalau nanti nambah
// handler baru (Tahap 5+), cukup tambah 1 field di sini -- tidak perlu
// ubah urutan argumen di pemanggilan SetupRoutes yang sudah ada.
type RouteDeps struct {
	AuthHandler       *AuthHandler
	NoteHandler       *NoteHandler
	NoteAccessHandler *NoteAccessHandler
	BookmarkHandler   *VideoBookmarkHandler
	AttachmentHandler *AttachmentHandler
	LikeHandler       *LikeHandler
	SaveHandler       *SaveHandler
	FollowHandler     *FollowHandler
	FeedHandler       *FeedHandler
	PlaylistHandler   *PlaylistHandler

	NoteRepo       repository.NoteRepository
	NoteAccessRepo repository.NoteAccessRepository

	JWTSecret string
}

// SetupRoutes mendaftarkan semua route API. Fungsi ini yang jadi "peta"
// URL -> handler mana yang menangani. Sengaja dipisah dari main.go supaya
// main.go tetap tipis (cuma wiring).
func SetupRoutes(
	app *fiber.App,
	d RouteDeps,
) {
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", d.AuthHandler.Register)
	auth.Post("/login", d.AuthHandler.Login)

	// FYP -- Public, boleh diakses tanpa login (OptionalAuthMiddleware
	// dipakai supaya filter "following" bisa jalan KALAU kebetulan login).
	api.Get("/fyp",
		middleware.OptionalAuthMiddleware(d.JWTSecret),
		d.FeedHandler.GetFYP,
	)

	// Note routes
	notes := api.Group("/notes")
	notes.Post("/", middleware.AuthMiddleware(d.JWTSecret), d.NoteHandler.Create)
	notes.Get("/", middleware.AuthMiddleware(d.JWTSecret), d.NoteHandler.ListMine)
	// /notes/saved didaftarkan SEBELUM /notes/:slug sebagai kebiasaan
	// aman: rute statis ("saved") berpotensi tabrakan dengan rute
	// berparameter (:slug) yang polanya sama-sama 1 segment path --
	// mendaftarkan yang statis lebih dulu adalah praktik defensif umum
	// di banyak router, terlepas dari detail prioritas internal Fiber.
	notes.Get("/saved", middleware.AuthMiddleware(d.JWTSecret), d.SaveHandler.ListSaved)

	notes.Put("/:id",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteHandler.Update,
	)
	notes.Delete("/:id",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteHandler.Delete,
	)

	notes.Get("/:slug",
		middleware.OptionalAuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteHandler.GetBySlug,
	)

	// --- Note Access (Owner Only): kasih/ubah/cabut/lihat akses viewer
	// atau editor ke user lain lewat email. AccessMiddleware dipasang di
	// sini BUKAN untuk membatasi Owner-only-nya (itu sudah dicek ulang
	// sendiri di NoteAccessUsecase, independen dari middleware), tapi
	// supaya note-nya ke-resolve dan tersedia lewat c.Locals("note").
	notes.Post("/:id/access",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteAccessHandler.Grant,
	)
	notes.Delete("/:id/access",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteAccessHandler.Revoke,
	)
	notes.Get("/:id/access",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.NoteAccessHandler.List,
	)

	// --- Video Bookmark routes
	notes.Post("/:id/bookmarks",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.BookmarkHandler.Add,
	)
	notes.Get("/:id/bookmarks",
		middleware.OptionalAuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.BookmarkHandler.List,
	)
	notes.Delete("/:id/bookmarks/:bookmarkId",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.BookmarkHandler.Remove,
	)

	notes.Post("/:id/upload",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.AttachmentHandler.Upload,
	)
	notes.Get("/:id/attachments",
		middleware.OptionalAuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.AttachmentHandler.List,
	)
	notes.Delete("/:id/attachments/:attachmentId",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.AttachmentHandler.Remove,
	)

	// --- Like & Save (Toggle). Keduanya wajib login + note-nya harus
	// bisa diakses (AccessMiddleware) -- tidak masuk akal like/save note
	// yang bahkan tidak boleh kamu lihat.
	notes.Post("/:id/like",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.LikeHandler.Toggle,
	)
	notes.Post("/:id/save",
		middleware.AuthMiddleware(d.JWTSecret),
		middleware.AccessMiddleware(d.NoteRepo, d.NoteAccessRepo),
		d.SaveHandler.Toggle,
	)

	// --- Follow (Toggle). TIDAK lewat AccessMiddleware -- targetnya USER,
	// bukan note, jadi tidak ada urusan dengan visibility/note_access.
	users := api.Group("/users")
	users.Post("/:id/follow", middleware.AuthMiddleware(d.JWTSecret), d.FollowHandler.Toggle)

	// --- Playlist. TIDAK memakai AccessMiddleware sama sekali (itu
	// khusus untuk resource Note) -- otorisasi playlist (owner-only untuk
	// mutasi, private/public untuk GetNotes) ditangani LANGSUNG di dalam
	// PlaylistUsecase, karena modelnya lebih sederhana (tidak ada role
	// editor/viewer terpisah kayak note_access).
	playlists := api.Group("/playlists")
	playlists.Post("/", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.Create)
	playlists.Get("/", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.ListMine)
	playlists.Put("/:id", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.Update)
	playlists.Delete("/:id", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.Delete)
	playlists.Post("/:id/notes", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.AddNote)
	playlists.Delete("/:id/notes/:noteId", middleware.AuthMiddleware(d.JWTSecret), d.PlaylistHandler.RemoveNote)
	// GetNotes pakai OptionalAuthMiddleware -- playlist public boleh
	// dilihat guest, playlist private butuh login DAN harus pemiliknya
	// (dicek di usecase).
	playlists.Get("/:id", middleware.OptionalAuthMiddleware(d.JWTSecret), d.PlaylistHandler.GetNotes)

}
