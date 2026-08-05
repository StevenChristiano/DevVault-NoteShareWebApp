package usecase

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
)

type SaveUsecase interface {
	Toggle(noteID, userID uint) (saved bool, err error)
	ListSaved(userID uint) ([]entity.Note, error)
}

type saveUsecase struct {
	savedRepo repository.SavedNoteRepository
}

func NewSaveUsecase(savedRepo repository.SavedNoteRepository) SaveUsecase {
	return &saveUsecase{savedRepo: savedRepo}
}

func (u *saveUsecase) Toggle(noteID, userID uint) (bool, error) {
	existing, err := u.savedRepo.FindByUserAndNote(userID, noteID)
	if err != nil {
		return false, err
	}

	if existing != nil {
		if err := u.savedRepo.Delete(userID, noteID); err != nil {
			return false, err
		}
		return false, nil //false == unsaved
	}

	if err := u.savedRepo.Create(userID, noteID); err != nil {
		return false, err
	}
	return true, nil //true == saved
}

// ListSaved meng-"ratakan" []entity.SavedNote jadi []entity.Note --
// pemanggil (handler) cuma peduli soal note-nya, tidak perlu tahu ada
// struct perantara SavedNote di baliknya.
func (u *saveUsecase) ListSaved(userID uint) ([]entity.Note, error) {
	savedList, err := u.savedRepo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	notes := make([]entity.Note, 0, len(savedList))
	for _, s := range savedList {
		notes = append(notes, s.Note)
	}
	return notes, nil
}
