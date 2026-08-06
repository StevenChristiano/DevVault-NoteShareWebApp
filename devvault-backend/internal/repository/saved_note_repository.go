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
	// ListByUserID meng-Preload Note, supaya usecase langsung dapat data
	// note lengkap (header, slug, dst) tanpa query tambahan satu-satu --
	// dipakai endpoint GET /notes/saved.
	ListByUserID(userID uint) ([]entity.SavedNote, error)
	// GetOrCreate: kalau user SUDAH pernah save note ini, kembalikan baris
	// yang sudah ada (dengan ID-nya). Kalau belum, BUAT baru lalu
	// kembalikan. Dipakai fitur Playlist: "masukin note ke playlist"
	// otomatis men-save note itu juga kalau belum ke-save -- menjamin
	// invariant "semua note di playlist manapun pasti ada di Recently
	// Saved" TANPA memaksa user save manual dulu sebagai langkah terpisah.
	GetOrCreate(userID, noteID uint) (*entity.SavedNote, error)
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

func (r *savedNoteRepository) GetOrCreate(userID, noteID uint) (*entity.SavedNote, error) {
	existing, err := r.FindByUserAndNote(userID, noteID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	newSaved := &entity.SavedNote{UserID: userID, NoteID: noteID}
	if err := r.db.Create(newSaved).Error; err != nil {
		return nil, err
	}
	return newSaved, nil
}
