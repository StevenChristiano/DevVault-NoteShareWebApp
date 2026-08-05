package repository


import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

type SavedNoteRepository interface {
	FindByUserAndNote(userID, noteID uint) (*entity.SavedNote, error)
	Create(userID, noteID uint) error
	Delete(userID, noteID uint) error
	ListByUserID(userID uint) ([]entity. SavedNote, error)
}

type savedNoteRepository struct {
	db *gorm.DB
}

func NewSavedNoteRepository(db *gorm.DB) SavedNoteRepository {
	return &savedNoteRepository{db: db}
}

func (r *savedNoteRepository) FindByUserAndNote(userID, noteID uint) (*entity.SavedNote, error) {
	var saved entity.SavedNote
	err := r.db.Where("user_id = ? AND note_id = ?", userID, noteID).First(&saved).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &saved, nil
}

func (r *savedNoteRepository) Create(userID, noteID uint) error {
	return r.db.Create(&entity.SavedNote{UserID: userID, NoteID: noteID}).Error
}

func (r *savedNoteRepository) Delete(userID, noteID uint) error {
	return r.db.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&entity.SavedNote{}).Error
}

func (r *savedNoteRepository) ListByUserID(userID uint) ([]entity.SavedNote, error) {
	var saved []entity.SavedNote
	err := r.db.Preload("Note").Where("user_id = ?", userID).Order("created_at DESC").Find(&saved).Error
	if err != nil {
		return nil, err
	}
	return saved, nil
}