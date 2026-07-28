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
	FindByID(id uint) (*entity.Note, error)
	FindBySlug(slug string) (*entity.Note, error)
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