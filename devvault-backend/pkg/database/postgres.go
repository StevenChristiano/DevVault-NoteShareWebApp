package database

import (
	"fmt"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect membuka koneksi GORM ke PostgreSQL berdasarkan DatabaseConfig.
// Mengembalikan *gorm.DB yang nanti di-"suntikkan" (dependency injection)
// ke repository layer di tahap berikutnya.
func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.DSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger.Info di development supaya kita BISA LIHAT query SQL asli
		// yang di-generate GORM di terminal — sangat membantu buat belajar
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("gagal konek ke database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal ambil underlying sql.DB: %w", err)
	}

	// Ping memastikan koneksi BENAR-BENAR hidup, bukan cuma berhasil
	// membuat objek koneksi (gorm.Open bisa "berhasil" walau DB-nya mati,
	// errornya baru ketahuan saat query pertama tanpa Ping ini).
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	return db, nil
}
