// cmd/api/main.go adalah SATU-SATUNYA entrypoint aplikasi backend.
// Isinya sengaja tipis: cuma "merangkai" (wiring) potongan-potongan dari
// package lain (config, database) lalu menjalankan server. Logika bisnis
// TIDAK ditulis di sini — nanti bakal tinggal di internal/usecase dan
// internal/delivery/http (Tahap 2 dst).
package main

import (
	"log"

	"github.com/StevenChristiano/devvault-backend/internal/config"
	"github.com/StevenChristiano/devvault-backend/pkg/database"
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

	// 3. Auto-migrate 8 tabel sesuai skema di dokumen teknis.
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("gagal migrasi database: %v", err)
	}
	log.Println("✅ auto-migrate selesai, 8 tabel siap")

	// 4. Setup Fiber app. Untuk Tahap 1 kita cuma buat satu endpoint
	//    health-check, sekadar bukti server hidup dan bisa dites lewat
	//    browser/Postman. Endpoint asli (register, login, notes, dst)
	//    baru masuk mulai Tahap 2 & 3, ditaruh di internal/delivery/http.
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.App.Env,
		})
	})

	addr := ":" + cfg.App.Port
	log.Printf("🚀 server jalan di http://localhost%s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server berhenti: %v", err)
	}
}
