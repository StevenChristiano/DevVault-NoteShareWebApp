package database

import (
	"fmt"

	"github.com/StevenChristiano/devvault-backend/internal/entity"
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