package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

type AttachmentRepository interface {
	Create(attachment *entity.Attachment) error
	FindByID(id uint) (*entity.Attachment, error)
	Delete(id uint) error
	ListByNoteID(noteID uint) ([]entity.Attachment, error)
}

type attachmentRepository struct {
	db *gorm.DB
}

func NewAttachmentRepository(db *gorm.DB) AttachmentRepository {
	return &attachmentRepository{db: db}
}

func (r *attachmentRepository) Create(attachment *entity.Attachment) error {
	return r.db.Create(attachment).Error
}

func (r *attachmentRepository) FindByID(id uint) (*entity.Attachment, error) {
	var attachment entity.Attachment
	err := r.db.First(&attachment, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attachment, nil
}

func (r *attachmentRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Attachment{}, id).Error
}

func (r *attachmentRepository) ListByNoteID(noteID uint) ([]entity.Attachment, error) {
	var attachments []entity.Attachment
	err := r.db.Where("note_id = ?", noteID).Order("created_at DESC").Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}
