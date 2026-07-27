package entity

import "time"

/*
User merepresentasikan tabel `user` di database.

CATATAN PENTING dari dokumen spesifikasi:
User TIDAK punya kolom "role" global. Peran (Owner/Editor/Viewer)
selalu ditentukan kontekstual per note (lihat NoteAccess), bukan
melekat permanen ke akun user. Makanya struct ini sengaja polos,
tidak ada field Role di sini.
*/
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"type:varchar(50);not null"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	AvatarURL    string `gorm:"type:varchar(255)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

/*
AvatarURL: path/URL ke file foto profil (disimpan di disk/S3/Cloudinary,
bukan di database). SENGAJA nullable (tidak ada `not null`) — kalau
kosong, artinya user belum upload foto sendiri. Frontend yang
menentukan tampilannya sebagai avatar default (mis. lingkaran
inisial nama), backend tidak perlu tahu/menyimpan status "pakai
default atau tidak", cukup jawab apa adanya: kosong atau tidak.
*/
