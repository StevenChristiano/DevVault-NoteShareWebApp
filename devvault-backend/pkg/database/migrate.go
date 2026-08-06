package database

import (
	"fmt"
	"log"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

/* AutoMigrate membuat/menyesuaikan semua 8 tabel sesuai skema di dokumen
teknis. Urutan di slice ini SENGAJA disusun: tabel independen (User)
duluan, baru tabel yang punya foreign key ke tabel sebelumnya. GORM
sebenarnya cukup pintar menangani urutan FK sendiri, tapi menulis
urutan yang logis begini bikin kode lebih gampang dibaca manusia juga.

PENTING dipahami soal AutoMigrate (bukan golang-migrate/versioned SQL):
  - AutoMigrate akan MEMBUAT tabel yang belum ada, MENAMBAH kolom yang
  belum ada, dan menambah index/constraint yang belum ada.
	- AutoMigrate TIDAK PERNAH menghapus kolom atau mengubah tipe data
	kolom yang sudah ada (demi keamanan data). Kalau kamu ubah tipe
	field di struct Go, kamu harus ubah manual di database atau drop
  tabelnya sendiri saat development.
	- Ini pilihan yang cocok untuk project solo/portofolio: cepat, tanpa
  perlu menulis file .sql migration manual. Untuk tim/production yang
	butuh riwayat perubahan skema & kemampuan rollback, biasanya orang
	pindah ke tool seperti golang-migrate atau Atlas.
*/

func AutoMigrate(db *gorm.DB) error {
	// Migrasi manual DULU (kalau perlu), baru AutoMigrate menangani
	// sisanya (bikin tabel baru, nambah kolom baru yang belum ada, dst).
	if err := migrateSavedNotesPrimaryKey(db); err != nil {
		return fmt.Errorf("migrasi manual saved_notes gagal: %w", err)
	}

	models := []interface{} {
		&entity.User{},
		&entity.Note{},
		&entity.NoteAccess{},
		&entity.Like{},
		&entity.SavedNote{},
		&entity.Follow{},
		&entity.Attachment{},
		&entity.VideoBookmark{},
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate gagal: %w", err)
	}

	return nil
}

// migrateSavedNotesPrimaryKey mengubah primary key tabel saved_notes dari
// composite (user_id, note_id) menjadi surrogate `id` -- dibutuhkan
// karena tabel baru playlist_notes butuh foreign key ke SATU baris
// spesifik saved_notes (bukan ke kombinasi 2 kolom).
//
// IDEMPOTENT: fungsi ini AMAN dipanggil berkali-kali (tiap kali server
// start). Dia cuma benar-benar menjalankan ALTER TABLE kalau mendeteksi
// migrasi ini BELUM pernah dilakukan (kolom `id` belum ada) DAN tabelnya
// memang sudah ada sebelumnya (database yang benar-benar baru tidak
// perlu migrasi ini -- AutoMigrate di bawah akan langsung membuat tabel
// dengan struct BARU dari awal).
func migrateSavedNotesPrimaryKey(db *gorm.DB) error {
	var tableExists bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.tables WHERE table_name = 'saved_notes'
	)`).Scan(&tableExists).Error; err != nil {
		return err
	}
	if !tableExists {
		return nil // database baru, AutoMigrate akan bikin dari struct terbaru langsung
	}

	var hasIDColumn bool
	if err := db.Raw(`SELECT EXISTS (
		SELECT 1 FROM information_schema.columns
		WHERE table_name = 'saved_notes' AND column_name = 'id'
	)`).Scan(&hasIDColumn).Error; err != nil {
		return err
	}
	if hasIDColumn {
		return nil // sudah pernah dimigrasi sebelumnya, tidak perlu diulang
	}

	log.Println("[migrate] mengubah primary key saved_notes dari composite (user_id, note_id) ke surrogate id...")

	statements := []string{
		// Nama constraint primary key lama mengikuti konvensi default
		// Postgres: "{nama_tabel}_pkey". IF EXISTS berjaga-jaga kalau
		// constraint-nya ternyata dinamai beda di komputer kamu.
		`ALTER TABLE saved_notes DROP CONSTRAINT IF EXISTS saved_notes_pkey`,
		`ALTER TABLE saved_notes ADD COLUMN id SERIAL PRIMARY KEY`,
		`ALTER TABLE saved_notes ADD CONSTRAINT idx_saved_notes_user_note UNIQUE (user_id, note_id)`,
	}
	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("gagal eksekusi %q: %w", stmt, err)
		}
	}

	log.Println("[migrate] migrasi saved_notes selesai — data lama TETAP UTUH, cuma primary key-nya berubah bentuk")
	return nil
}
