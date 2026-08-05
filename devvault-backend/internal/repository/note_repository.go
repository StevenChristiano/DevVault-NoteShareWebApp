package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

// NoteRepository di Tahap 2 SENGAJA baru berisi method baca (FindByID,
// FindBySlug) — cukup untuk kebutuhan AccessMiddleware mengecek
// visibility sebuah note. Method tulis (Create/Update/Delete) menyusul
// di Tahap 3 waktu kita bangun CRUD Note lengkap.
type NoteRepository interface {
	Create(note *entity.Note) error
	Update(note *entity.Note) error
	Delete(id uint) error
	FindByID(id uint) (*entity.Note, error)
	FindBySlug(slug string) (*entity.Note, error)
	FindByUserID(userID uint) ([]entity.Note, error)
	ExistsBySlug(slug string) (bool, error)
	IncrementViewCount(id uint) error
	// ListPublicFeed dipakai endpoint FYP. sortBy: "likes" | "saves" | "latest".
	// followerID: kalau tidak nil, feed DIFILTER cuma note dari user yang
	// di-follow oleh followerID (sesuai dokumen: "difilter khusus
	// menampilkan note dari user yang di-follow").
	ListPublicFeed(sortBy string, followerID *uint, limit, offset int) ([]entity.Note, error)
}

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) FindByID(id uint) (*entity.Note, error) {
	var note entity.Note
	err := r.db.First(&note, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) FindBySlug(slug string) (*entity.Note, error) {
	var note entity.Note
	err := r.db.Where("slug = ?", slug).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &note, nil
}

func (r *noteRepository) Create(note *entity.Note) error {
	return r.db.Create(note).Error
}

func (r *noteRepository) Update(note *entity.Note) error {
	return r.db.Save(note).Error
}

func (r *noteRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Note{}, id).Error
}

func (r *noteRepository) FindByUserID(userID uint) ([]entity.Note, error) {
	var notes []entity.Note
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *noteRepository) ExistsBySlug(slug string) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Note{}).Where("slug = ?", slug).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *noteRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&entity.Note{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

// ListPublicFeed. Catatan soal "saves" sorting: tidak ada kolom cache
// save_count di tabel notes (beda dengan like_count yang di-cache) --
// jadi diurutkan lewat SUBQUERY COUNT langsung di klausa ORDER BY.
// Ini valid di Postgres tanpa perlu men-SELECT hasil subquery-nya secara
// eksplisit. Performanya cukup untuk skala project ini; kalau nanti data
// membesar drastis, langkah optimasi lanjutannya adalah menambah kolom
// cache save_count (sama seperti like_count) -- sengaja tidak dilakukan
// sekarang mengikuti prinsip YAGNI yang sudah kita bahas sebelumnya.
func (r *noteRepository) ListPublicFeed(sortBy string, followerID *uint, limit, offset int) ([]entity.Note, error) {
	query := r.db.Model(&entity.Note{}).Where("visibility = ?", entity.VisibilityPublic)

	if followerID != nil {
		query = query.Where("user_id IN (SELECT following_id FROM follows WHERE follower_id = ?)", *followerID)
	}

	switch sortBy {
	case "likes":
		query = query.Order("like_count DESC")
	case "saves":
		query = query.Order("(SELECT COUNT(*) FROM saved_notes WHERE saved_notes.note_id = notes.id) DESC")
	default: // "latest" atau nilai lain yang tidak dikenal -> fallback aman
		query = query.Order("created_at DESC")
	}

	var notes []entity.Note
	err := query.Limit(limit).Offset(offset).Find(&notes).Error
	if err != nil {
		return nil, err
	}
	return notes, nil
}