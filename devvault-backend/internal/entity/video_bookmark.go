package entity

import "time"

// VideoBookmark merepresentasikan tabel `video_bookmark`.
// Video YouTube itu sendiri TIDAK disimpan di server kita — cuma
// referensi URL/ID-nya. Yang kita simpan adalah anotasi/catatan di
// detik tertentu (dipakai untuk player.seekTo() di frontend).
type VideoBookmark struct {
	ID uint `gorm:"primaryKey"`

	NoteID uint `gorm:"not null;index"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	YoutubeURL string `gorm:"type:varchar(255);not null"` // link lengkap yang ditempel user
	YoutubeID  string `gorm:"type:varchar(50);not null"`  // hasil parsing, mis. dQw4w9WgXcQ

	TimestampSec int    `gorm:"not null"` // detik ke-N di video
	NoteText     string `gorm:"type:text"`

	CreatedAt time.Time
}
