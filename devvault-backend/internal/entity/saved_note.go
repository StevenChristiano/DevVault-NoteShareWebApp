package entity

import "time"

// SavedNote merepresentasikan tabel `saved_notes`.
//
// Beda dengan Like/NoteAccess yang punya kolom `id` sendiri, tabel ini
// SENGAJA tidak punya id — primary key-nya adalah GABUNGAN (composite)
// dari user_id + note_id, sesuai dokumen ("Composite key bersama note_id").
//
// KENAPA composite PK di sini masuk akal:
// Secara alami, kombinasi (user, note) itu SENDIRI sudah unik dan itulah
// identitas barisnya — tidak butuh id buatan (surrogate key) karena tidak
// ada makna tambahan selain "user ini pernah save note ini". Beda dengan
// Like yang kita kasih id sendiri (lebih konsisten dengan gaya log/event
// yang mungkin di masa depan butuh direferensikan satu-per-satu).
type SavedNote struct {
	UserID uint `gorm:"primaryKey"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	NoteID uint `gorm:"primaryKey"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
}
