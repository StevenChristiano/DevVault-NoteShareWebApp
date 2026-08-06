// cmd/api/main.go adalah SATU-SATUNYA entrypoint aplikasi backend.
// Isinya sengaja tipis: cuma "merangkai" (wiring) potongan-potongan dari
// package lain (config, database, repository, usecase, delivery/http)
// lalu menjalankan server. Logika bisnis TIDAK ditulis di sini.
package main

import (
	"log"
	"time"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/config"
	deliveryhttp "github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/delivery/http"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/database"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/storage"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// 1. Load konfigurasi dari .env / environment variable.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("gagal load config: %v", err)
	}

	//2. Buka koneksi ke PostgreSQL.
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("gagal konek database: %v", err)
	}
	log.Println("✅ berhasil konek ke PostgreSQL")

	// 3. Auto-migrate skema database (10 tabel: 8 tabel awal + playlists,
	//    playlist_notes yang ditambahkan belakangan).
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("gagal migrasi database: %v", err)
	}
	log.Println("✅ auto-migrate selesai, 10 tabel siap")

	// 4. Rangkai (wire) repository -> usecase -> handler.
	//    Perhatikan arah panahnya: authHandler BUTUH authUsecase, authUsecase
	//    BUTUH userRepo. Jadi urutan pembuatannya HARUS dari yang paling
	//    "dalam" (repository) ke yang paling "luar" (handler) — sama
	//    persis dengan prinsip dependency graph yang sudah kita bahas.
	userRepo := repository.NewUserRepository(db)
	noteRepo := repository.NewNoteRepository(db)
	noteAccessRepo := repository.NewNoteAccessRepository(db)
	bookmarkRepo := repository.NewVideoBookmarkRepository(db)
	attachmentRepo := repository.NewAttachmentRepository(db)
	likeRepo := repository.NewLikeRepository(db)
	savedNoteRepo := repository.NewSavedNoteRepository(db)
	followRepo := repository.NewFollowRepository(db)
	playlistRepo := repository.NewPlaylistRepository(db)
	localStorage := storage.NewLocalStorage(cfg.App.UploadDir)

	jwtTTL := time.Duration(cfg.JWT.ExpiryHour) * time.Hour
	authUsecase := usecase.NewAuthUsecase(userRepo, cfg.JWT.Secret, jwtTTL)
	noteUsecase := usecase.NewNoteUsecase(noteRepo)
	noteAccessUsecase := usecase.NewNoteAccessUsecase(noteRepo, userRepo, noteAccessRepo)
	bookmarkUsecase := usecase.NewVideoBookmarkUsecase(bookmarkRepo)
	attachmentUsecase := usecase.NewAttachmentUsecase(attachmentRepo, localStorage)
	likeUsecase := usecase.NewLikeUsecase(likeRepo)
	saveUsecase := usecase.NewSaveUsecase(savedNoteRepo)
	followUsecase := usecase.NewFollowUsecase(followRepo)
	feedUsecase := usecase.NewFeedUsecase(noteRepo)
	playlistUsecase := usecase.NewPlaylistUsecase(playlistRepo, savedNoteRepo, noteRepo, noteAccessRepo)

	authHandler := deliveryhttp.NewAuthHandler(authUsecase)
	noteHandler := deliveryhttp.NewNoteHandler(noteUsecase)
	noteAccessHandler := deliveryhttp.NewNoteAccessHandler(noteAccessUsecase)
	bookmarkHandler := deliveryhttp.NewVideoBookmarkHandler(bookmarkUsecase)
	attachmentHandler := deliveryhttp.NewAttachmentHandler(attachmentUsecase)
	likeHandler := deliveryhttp.NewLikeHandler(likeUsecase)
	saveHandler := deliveryhttp.NewSaveHandler(saveUsecase)
	followHandler := deliveryhttp.NewFollowHandler(followUsecase)
	feedHandler := deliveryhttp.NewFeedHandler(feedUsecase)
	playlistHandler := deliveryhttp.NewPlaylistHandler(playlistUsecase)

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // 20 MB
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.App.Env,
		})
	})

	deliveryhttp.SetupRoutes(app, deliveryhttp.RouteDeps{
		AuthHandler:       authHandler,
		NoteHandler:       noteHandler,
		NoteAccessHandler: noteAccessHandler,
		BookmarkHandler:   bookmarkHandler,
		AttachmentHandler: attachmentHandler,
		LikeHandler:       likeHandler,
		SaveHandler:       saveHandler,
		FollowHandler:     followHandler,
		FeedHandler:       feedHandler,
		PlaylistHandler:   playlistHandler,
		NoteRepo:           noteRepo,
		NoteAccessRepo:     noteAccessRepo,
		JWTSecret:          cfg.JWT.Secret,
	})

	addr := ":" + cfg.App.Port
	log.Printf("🚀 server jalan di http://localhost%s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server berhenti: %v", err)
	}
}
