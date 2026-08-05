package repository


import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
)

type FollowRepository interface {
	FindByFollowerAndFollowing(followerID, followingID uint) (*entity.Follow, error)
	Create(followerID, followingID uint) error
	Delete(followerID, followingID uint) error
	ListFollowingIDs(followerID uint) ([]uint, error)
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) FindByFollowerAndFollowing(followerID, followingID uint) (*entity.Follow, error) {
	var follow entity.Follow
	err := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &follow, nil
}

func (r *followRepository) Create(followerID, followingID uint) error {
	return r.db.Create(&entity.Follow{FollowerID: followerID, FollowingID: followingID}).Error
}

func (r *followRepository) Delete(followerID, followingID uint) error {
	return r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&entity.Follow{}).Error
}

func (r *followRepository) ListFollowingIDs(followerID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&entity.Follow{}).Where("follower_id = ?", followerID).Pluck("following_id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}