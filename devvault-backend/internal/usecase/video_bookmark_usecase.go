package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/youtube"
)

var ErrBookmarkNotFound = errors.New("bookmark not found")

type VideoBookmarkUsecase interface {
	Add(noteID uint, role string, youtubeURL string, timestampSec int, noteText string) (*entity.VideoBookmark, error)
	List(noteID uint) ([]entity.VideoBookmark, error)
	Remove(noteID uint, role string, bookmarkID uint) error
}

type videoBookmarkUsecase struct {
	bookmarkRepo repository.VideoBookmarkRepository
}

func NewVideoBookmarkUsecase(bookmarkRepo repository.VideoBookmarkRepository) VideoBookmarkUsecase {
	return &videoBookmarkUsecase{bookmarkRepo: bookmarkRepo}
}

// canEdit itu helper kecil dipakai Add & Remove -- sengaja dipisah biar
// tidak menulis kondisi yang sama 2 kali (dan supaya kalau aturannya
// berubah nanti, cukup diubah di satu tempat).
func canEdit(role string) bool {
	return role == "owner" || role == "editor"
}

func (u *videoBookmarkUsecase) Add(noteID uint, role string, youtubeURL string, timestampSec int, noteText string) (*entity.VideoBookmark, error) {
	if !canEdit(role) {
		return nil, ErrForbidden
	}

	youtubeID, err := youtube.ExtractID(youtubeURL)
	if err != nil {
		return nil, err
	}

	bookmark := &entity.VideoBookmark{
		NoteID:       noteID,
		YoutubeURL:   youtubeURL,
		YoutubeID:    youtubeID,
		TimestampSec: timestampSec,
		NoteText:     noteText,
	}

	if err := u.bookmarkRepo.Create(bookmark); err != nil {
		return nil, err
	}
	return bookmark, nil
}

func (u *videoBookmarkUsecase) List(noteID uint) ([]entity.VideoBookmark, error) {
	return u.bookmarkRepo.ListByNoteID(noteID)
}

func (u *videoBookmarkUsecase) Remove(noteID uint, role string, bookmarkID uint) error {
	if !canEdit(role) {
		return ErrForbidden
	}

	bookmark, err := u.bookmarkRepo.FindByID(bookmarkID)
	if err != nil {
		return err
	}
	if bookmark == nil {
		return ErrBookmarkNotFound
	}
	if bookmark.NoteID != noteID { // to prevent deleting a bookmark that doesn't belong to the note
		return ErrBookmarkNotFound
	}

	return u.bookmarkRepo.Delete(bookmarkID)
}
