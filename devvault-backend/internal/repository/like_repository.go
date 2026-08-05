package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)


type LikeRepository interface {
	FindByNoteAndUser(noteID, userID uint) (*entity.Like, error)
	Create(noteID, userID uint) error
	Delete(noteID, userID uint) error
}

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) FindByNoteAndUser(noteID, userID uint) (*entity.Like, error) {
	var like entity.Like
	err := r.db.Where ("note_id = ? AND user_id = ?", noteID, userID).First(&like).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &like, nil
}

func (r *likeRepository) Create(noteID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entity.Like{NoteID: noteID, UserID: userID}).Error; err != nil {
			return err
		}
		return tx.Model(&entity.Note{}).Where("id = ?", noteID).
			UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
}

func (r *likeRepository) Delete(noteID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("note_id = ? AND user_id = ?", noteID, userID).Delete(&entity.Like{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Model(&entity.Note{}).Where("id = ?", noteID).
			UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error
	})
}