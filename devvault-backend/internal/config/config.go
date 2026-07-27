// Package config bertugas MEMBACA environment variable (.env) dan
// menyediakannya dalam bentuk struct Go yang gampang dipakai layer lain.
//
// KENAPA dipisah jadi package sendiri?
// Karena kalau nanti kita butuh config baru (misalnya JWT_SECRET di Tahap 2,
// atau UPLOAD_DIR di Tahap 3), kita cukup nambah field di sini — layer lain
// (database, handler, dst) tidak perlu tahu caranya baca env, cukup panggil
// config.Load().
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config menampung semua nilai konfigurasi aplikasi.
// Kita pisah per "kelompok" (Database, App) supaya rapi saat field-nya
// bertambah banyak di tahap-tahap berikutnya.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
}

type AppConfig struct {
	Port string // contoh: "8080"
	Env  string // "development" | "production"
}	

type DatabaseConfig struct {
	Host		 		string
	Port		 		string
	User		 		string
	Password		string
	Name		 		string
	SSLMode		 	string // "disable" untuk lokal, "require" untuk kebanyakan provider cloud
}

// Load membaca file .env (jika ada) lalu mengisi struct Config dari
// environment variable. Dipanggil sekali di main.go saat aplikasi start.
func Load() (*Config, error) {
	// godotenv.Load TIDAK error kalau file .env tidak ditemukan, hanya warning.
	// Ini penting supaya di server produksi (yang biasanya set env lewat
	// platform, bukan file .env) aplikasi tetap jalan normal.
	if err := godotenv.Load(); err != nil {
		fmt.Println("[config] .env tidak ditemukan, lanjut pakai environment variable OS")
	}
	
	cfg := &Config {
		App: AppConfig {
			Port: getEnv("APP_PORT", "8080"),
			Env: getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig {
			Host: getEnv("DB_HOST", "localhost"),
			Port: getEnv("DB_PORT", "5432"),
			User: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "grilleyz_1707"),
			Name: getEnv("DB_NAME", "devvault_db"),
			SSLMode: getEnv("DB_SSLMODE", "disable"),
		},
	}

	return cfg, nil
}

// DSN membangun connection string PostgreSQL dari config.
// Dipisah jadi method sendiri supaya package database (pkg/database) tidak
// perlu tahu format string DSN-nya seperti apa — cukup panggil cfg.Database.DSN().
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// getEnv adalah helper kecil: ambil env var, kalau kosong pakai default.
func getEnv (key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}