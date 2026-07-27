package entity

import "time"

// Role adalah peran yang diberikan Owner ke user lain pada note tertentu.
// Sengaja dibuat type terpisah dari Visibility walau sama-sama string,
// supaya kompiler Go mencegah kita salah taruh (mis. tidak sengaja isi
// field Role dengan "public" yang harusnya nilai Visibility).
type Role string

const (
	RoleViewer Role = "viewer"
	RoleEditor Role = "editor"
)

// NoteAccess merepresentasikan tabel `note_access`.
// Satu baris = satu kombinasi (note, user yang diberi akses) + peran.
// HANYA diisi kalau note.visibility = shared (aturan ini ditegakkan di
// usecase layer nanti, bukan di database — GORM/Postgres tidak tahu
// "kalau row ini ada maka note tsb wajib shared").
type NoteAccess struct {
	ID uint `gorm:"primaryKey"`

	NoteID uint `gorm:"not null;index:idx_note_user,unique"`
	Note   Note `gorm:"foreignKey:NoteID;references:ID;constraint:OnDelete:CASCADE"`

	UserID uint `gorm:"not null;index:idx_note_user,unique"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`

	Role Role `gorm:"type:varchar(20);not null;check:role IN ('viewer','editor')"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

