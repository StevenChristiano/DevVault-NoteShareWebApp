package entity

import "time"

// Like merepresentasikan tabel `likes`.
// Satu baris = satu like. Unlike dilakukan dengan MENGHAPUS baris ini,
// bukan menambah kolom `is_liked bool` — ini pilihan sadar dari dokumen
// spesifikasi (bukan keputusan kita), karena riwayat "kapan like diberikan"
// (created_at) juga berguna, dan hitungan gampang: COUNT(*) baris di sini.
type Like struct {
	ID uint `gorm:"primaryKey"`

	NoteID uint `gorm:"not null;index:idx_like_note_user,unique"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	UserID uint `gorm:"not null;index:idx_like_note_user,unique"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	// Hanya created_at (tidak ada updated_at) karena baris ini bersifat
	// log/event: sekali dibuat, tidak pernah di-UPDATE, hanya dibuat atau
	// dihapus. Ini sesuai konvensi umum yang disebutkan di dokumen teknis.
	CreatedAt time.Time
}
