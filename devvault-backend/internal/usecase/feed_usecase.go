package usecase

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
)

const defaultFeedLimit = 20

type FeedUsecase interface {
	// GetFYP. userID boleh nil (guest, sesuai dokumen "FYP Public").
	// followingOnly HANYA berlaku kalau userID tidak nil -- guest tidak
	// mungkin punya daftar following.
	GetFYP(sortBy string, userID *uint, followingOnly bool, page int) ([]entity.Note, error)
}

type feedUsecase struct {
	noteRepo repository.NoteRepository
}

func NewFeedUsecase(noteRepo repository.NoteRepository) FeedUsecase {
	return &feedUsecase{noteRepo: noteRepo}
}

func (u *feedUsecase) GetFYP(sortBy string, userID *uint, followingOnly bool, page int) ([]entity.Note, error) {
	var followerID *uint
	if followingOnly && userID != nil {
		followerID = userID
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * defaultFeedLimit

	return u.noteRepo.ListPublicFeed(sortBy, followerID, defaultFeedLimit, offset)
}
