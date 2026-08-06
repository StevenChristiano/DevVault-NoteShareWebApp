package entity

import "time"

// SavedNote merepresentasikan tabel `saved_notes`.
//
// UPDATE (Tahap 4 lanjutan, fitur Playlist): tabel ini SEKARANG punya
// kolom `id` sendiri -- berubah dari keputusan awal (composite PK
// user_id+note_id tanpa id). Ini PERSIS skenario yang sudah diantisipasi
// sejak awal: begitu ada tabel LAIN (playlist_notes) yang butuh
// menunjuk ke baris SPESIFIK di sini (supaya kalau baris ini dihapus,
// keanggotaan di semua playlist ikut terhapus otomatis lewat FOREIGN
// KEY CASCADE), surrogate id jadi diperlukan.
//
// Keunikan (user_id, note_id) TETAP dijaga -- sekarang lewat
// uniqueIndex eksplisit, bukan lewat status primary key lagi.
//
// PENTING: karena ini perubahan PRIMARY KEY (bukan sekadar tambah kolom
// biasa), AutoMigrate TIDAK BISA melakukan ini otomatis untuk database
// yang SUDAH punya tabel saved_notes versi lama. Lihat
// pkg/database/migrate.go, fungsi migrateSavedNotesPrimaryKey -- migrasi
// manual (raw SQL) dijalankan otomatis SEKALI sebelum AutoMigrate,
// idempotent (aman dijalankan berkali-kali, cuma benar-benar migrasi
// kalau memang belum pernah).
type SavedNote struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	UserID uint `gorm:"not null;uniqueIndex:idx_saved_notes_user_note"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	NoteID uint `gorm:"not null;uniqueIndex:idx_saved_notes_user_note"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
}
