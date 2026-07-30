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