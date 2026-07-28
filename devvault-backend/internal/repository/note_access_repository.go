package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

// NoteAccessRepository dipakai AccessMiddleware untuk jawab pertanyaan:
// "untuk note yang visibility-nya shared, apakah user ini dikasih akses
// (viewer/editor) oleh Owner-nya?"
type NoteAccessRepository interface {
	FindByNoteAndUser(noteID, userID uint) (*entity.NoteAccess, error)
}

type noteAccessRepository struct {
	db *gorm.DB
}

func NewNoteAccessRepository(db *gorm.DB) NoteAccessRepository {
	return &noteAccessRepository{db: db}
}

func (r *noteAccessRepository) FindByNoteAndUser(noteID, userID uint) (*entity.NoteAccess, error) {
	var access entity.NoteAccess
	err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&access).Error
	if err != nil {
		if errors.Is(err, gorm.ErrCheckConstraintViolated) {
			return nil, nil
		}
		return nil, err
	}
	return &access, nil
}
