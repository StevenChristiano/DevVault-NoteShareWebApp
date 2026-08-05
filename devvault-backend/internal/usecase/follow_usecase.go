package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
)

var ErrCannotFollowSelf = errors.New("you cannot follow yourself")

type FollowUsecase interface {
	Toggle(followerID, followingID uint) (following bool, err error)
}

type followUsecase struct {
	followRepo repository.FollowRepository
}

func NewFollowUsecase(followRepo repository.FollowRepository) FollowUsecase {
	return &followUsecase{followRepo: followRepo}
}

func (u *followUsecase) Toggle(followerID, followingID uint) (bool, error) {
	if followerID == followingID {
		return false, ErrCannotFollowSelf
	}

	existing, err := u.followRepo.FindByFollowerAndFollowing(followerID, followingID)
	if err != nil {
		return false, err
	}

	if existing != nil {
		if err := u.followRepo.Delete(followerID, followingID); err != nil {
			return false, err
		}
		return false, nil // false == unfollowed
	}

	if err := u.followRepo.Create(followerID, followingID); err != nil {
		return false, err
	}
	return true, nil // true == followed
}
