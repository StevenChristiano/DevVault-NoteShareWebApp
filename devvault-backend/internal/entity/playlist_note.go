package entity

import "time"

// PlaylistNote merepresentasikan tabel `playlist_notes` -- relasi
// many-to-many antara Playlist dan SavedNote (BUKAN langsung ke Note).
//
// KENAPA menunjuk ke SavedNote, bukan Note langsung? Ini kuncinya:
// dengan foreign key ke SavedNote.ID + ON DELETE CASCADE, begitu user
// unsave sebuah note (baris di saved_notes terhapus), SEMUA baris
// playlist_notes yang menunjuk ke situ IKUT TERHAPUS OTOMATIS oleh
// database -- menjamin invariant "note yang ada di playlist manapun
// PASTI ada di Recently Saved" tanpa mengandalkan kode aplikasi untuk
// selalu ingat membersihkan playlist_notes secara manual tiap kali ada
// unsave (yang rawan lupa/human error kalau ditangani cuma di level Go).
type PlaylistNote struct {
	PlaylistID uint     `gorm:"primaryKey"`
	Playlist   Playlist `gorm:"foreignKey:PlaylistID;references:ID;constraint:OnDelete:CASCADE"`

	SavedNoteID uint      `gorm:"primaryKey"`
	SavedNote   SavedNote `gorm:"foreignKey:SavedNoteID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
}