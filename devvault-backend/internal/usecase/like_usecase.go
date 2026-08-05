package usecase

import "github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"

type LikeUsecase interface {
	Toggle(noteID, userID uint) (liked bool, err error) // switch like status: like == true, unlike == false
}

type likeUsecase struct {
	likeRepo repository.LikeRepository
}

func NewLikeUsecase(likeRepo repository.LikeRepository) LikeUsecase {
	return &likeUsecase{likeRepo: likeRepo}
}

func (u *likeUsecase) Toggle(noteID, userID uint) (bool, error) {
	existing, err := u.likeRepo.FindByNoteAndUser(noteID, userID)
	if err != nil {
		return false, err
	}

	if existing != nil {
		if err := u.likeRepo.Delete(noteID, userID); err != nil {
			return false, err
		}
		return false, nil // unlike
	}

	if err := u.likeRepo.Create(noteID, userID); err != nil {
		return false, err
	}
	return true, nil // like
}
