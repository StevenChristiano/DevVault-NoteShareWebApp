package entity

import "time"

// Playlist merepresentasikan tabel `playlists` -- fitur TAMBAHAN di luar
// 8 tabel final dokumen spek awal, buat mengelompokkan saved notes ke
// koleksi custom bernama (mirip playlist Spotify), terpisah dari
// "Recently Saved" (yang tetap murni dari tabel saved_notes, urut waktu).
//
// Visibility di sini SENGAJA cuma "private"/"public" (TIDAK ada "shared")
// -- beda dari Note yang punya 3 opsi. Keputusan sadar: playlist shared
// (invite viewer/editor spesifik per playlist) ditunda dulu, supaya tidak
// nambah 1 sistem akses baru yang mirip note_access tapi terpisah.
type Playlist struct {
	ID uint `gorm:"primaryKey"`

	UserID uint `gorm:"not null;index"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	Name string `gorm:"type:varchar(100);not null"`

	Visibility Visibility `gorm:"type:varchar(20);not null;default:private;check:visibility IN ('private','public')"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
