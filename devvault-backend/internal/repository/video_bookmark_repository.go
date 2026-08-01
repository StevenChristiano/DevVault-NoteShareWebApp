package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

type VideoBookmarkRepository interface {
	Create(bookmark *entity.VideoBookmark) error
	FindByID(id uint) (*entity.VideoBookmark, error)
	Delete(id uint) error
	ListByNoteID(noteID uint) ([]entity.VideoBookmark, error)
}

type videoBookmarkRepository struct {
	db *gorm.DB
}

func NewVideoBookmarkRepository(db *gorm.DB) VideoBookmarkRepository {
	return &videoBookmarkRepository{db: db}
}

func (r *videoBookmarkRepository) Create(bookmark *entity.VideoBookmark) error {
	return r.db.Create(bookmark).Error
}

func (r *videoBookmarkRepository) FindByID(id uint) (*entity.VideoBookmark, error) {
	var bookmark entity.VideoBookmark
	err := r.db.First(&bookmark, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &bookmark, nil
}

func (r *videoBookmarkRepository) Delete(id uint) error {
	return r.db.Delete(&entity.VideoBookmark{}, id).Error
}

func (r *videoBookmarkRepository) ListByNoteID(noteID uint) ([]entity.VideoBookmark, error) {
	var bookmarks []entity.VideoBookmark
	err := r.db.Where("note_id = ?", noteID).Order("timestamp_sec ASC").Find(&bookmarks).Error
	if err != nil {
		return nil, err
	}
	return bookmarks, nil
}
