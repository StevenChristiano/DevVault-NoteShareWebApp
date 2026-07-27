package entity

import "time"

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityShared  Visibility = "shared"
	VisibilityPublic  Visibility = "public"
)

// Note merepresentasikan tabel `note`.
type Note struct {
	ID     uint   `gorm:"primaryKey"`
	Header string `gorm:"type:varchar(255);not null"`
	Thumbnail string `gorm:"type:varchar(255)"`
	UserID uint `gorm:"not null;index"`
	User   User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Visibility Visibility `gorm:"type:varchar(20);not null;default:private;check:visibility IN ('private','shared','public')"`
	Slug string `gorm:"type:varchar(255);uniqueIndex"`
	Paragraph string `gorm:"type:text"`
	ViewCount int `gorm:"not null;default:0"`
	LikeCount int `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

/*
	Visibility adalah custom type untuk kolom note.visibility.
	KEPUTUSAN DESAIN: kenapa bukan native PostgreSQL ENUM type?

	Opsi A - Native Postgres ENUM (CREATE TYPE visibility_enum AS ENUM (...)):
		(+) Validasi nilai dijamin di level database, tidak mungkin ada typo.
		(-) GORM AutoMigrate TIDAK bisa membuat/mengubah native enum type secara
				otomatis (harus raw SQL manual di luar AutoMigrate).
		(-) Menambah nilai baru ke enum (mis. nanti butuh "archived") butuh
				ALTER TYPE ... ADD VALUE, yang di Postgres versi lama tidak bisa
				dijalankan di dalam transaction.

	Opsi B - varchar + CHECK constraint (yang dipakai di sini):
		(+) Kompatibel penuh dengan GORM AutoMigrate (tinggal tambah tag `check`).
		(+) Gampang ditambah nilai baru: tinggal ubah constraint & konstanta Go.
		(-) Validasi "sebenarnya" ada 2 lapis: constraint DB (jaring pengaman)
				dan validasi di usecase layer (baris pertahanan utama) — sedikit
				duplikasi tapi ini wajar dan disengaja (defense in depth).

	Untuk project belajar + AutoMigrate seperti ini, Opsi B jauh lebih
	bersahabat. Makanya dipakai di sini.
*/

/*
	- Thumbnail hanya menyimpan PATH string, bukan file binary.
	File aslinya nanti duduk di disk/storage (Tahap 3), di sini cuma
	referensinya. Kalau kosong, frontend yang menampilkan placeholder.

	- user_id adalah Owner. CASCADE artinya: kalau baris User dihapus,
	semua Note miliknya ikut terhapus otomatis oleh database (bukan
	oleh kode Go kita). Alternatifnya adalah RESTRICT (menolak hapus
	User selama masih punya Note) — dipilih CASCADE di sini karena
	untuk aplikasi personal knowledge base, note "yatim piatu" tanpa
	owner tidak ada gunanya.

	- Slug hanya terisi kalau Visibility = public (logika ini ada di
	usecase layer nanti, bukan di sini — entity cuma definisi struktur).

	- Paragraph menyimpan isi tulisan (markdown/rich text) TERMASUK
	penanda referensi seperti [[attachment:57]] atau [[bookmark:12]].
	`text` dipakai (bukan varchar) karena isinya bisa sangat panjang.

	- LikeCount adalah CACHE, bukan sumber kebenaran. Sumber kebenaran
	tetap tabel `likes` (dihitung COUNT saat like/unlike). Kolom ini
	cuma supaya query "tampilkan list note" tidak perlu JOIN + COUNT
	tabel likes tiap kali render — lebih cepat, trade-off-nya adalah
	kita HARUS ingat sinkronisasi manual di usecase Like (Tahap 4).
*/
