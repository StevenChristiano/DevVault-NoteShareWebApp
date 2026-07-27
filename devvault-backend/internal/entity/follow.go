package entity

import "time"

// Follow merepresentasikan tabel `follow` — relasi self-referencing
// many-to-many pada tabel User (satu user bisa follow banyak user,
// dan di-follow oleh banyak user).
//
// Kenapa dua foreign key sama-sama menunjuk ke User? Karena "follower"
// dan "following" itu sama-sama User, cuma berbeda peran dalam satu baris.
// Ini pola umum untuk relasi self-referencing (mirip tabel "friendship").
type Follow struct {
	FollowerID uint `gorm:"primaryKey"`
	Follower   User `gorm:"foreignKey:FollowerID;references:ID;constraint:OnDelete:CASCADE"`

	FollowingID uint `gorm:"primaryKey"`
	Following   User `gorm:"foreignKey:FollowingID;references:ID;constraint:OnDelete:CASCADE"`

	CreatedAt time.Time
}
