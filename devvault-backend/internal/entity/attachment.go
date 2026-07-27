package entity

import "time"

// Attachment merepresentasikan tabel `attachments`.
// ID di tabel ini dipakai sebagai penanda referensi di note.paragraph,
// misalnya [[attachment:57]] merujuk ke Attachment dengan ID 57.
type Attachment struct {
	ID uint `gorm:"primaryKey"`

	NoteID uint `gorm:"not null;index"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	FileName string `gorm:"type:varchar(255);not null"` // nama asli, ditampilkan ke user
	FilePath string `gorm:"type:varchar(255);not null"` // path fisik, sudah di-rename UUID
	FileType string `gorm:"type:varchar(100);not null"` // MIME type, mis. application/pdf

	CreatedAt time.Time
}
