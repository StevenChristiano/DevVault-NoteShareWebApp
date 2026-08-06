package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
)

var (
	ErrPlaylistNotFound    = errors.New("playlist not found")
	ErrInvalidPlaylistName = errors.New("playlist name is required")
	ErrInvalidVisibility   = errors.New("visibility must be either 'private' or 'public'")
)

type CreatePlaylistInput struct {
	Name       string
	Visibility entity.Visibility
}

type UpdatePlaylistInput struct {
	Name       *string
	Visibility *entity.Visibility
}

type PlaylistUsecase interface {
	Create(userID uint, input CreatePlaylistInput) (*entity.Playlist, error)
	ListMine(userID uint) ([]entity.Playlist, error)
	Update(userID, playlistID uint, input UpdatePlaylistInput) (*entity.Playlist, error)
	Delete(userID, playlistID uint) error
	// AddNote & RemoveNote: userID WAJIB pemilik playlist (cuma pemilik
	// yang boleh mengelola isi playlistnya sendiri).
	AddNote(userID, playlistID, noteID uint) error
	RemoveNote(userID, playlistID, noteID uint) error
	// GetNotes: requesterID BOLEH nil (guest), karena playlist bisa public.
	GetNotes(requesterID *uint, playlistID uint) (*entity.Playlist, []entity.Note, error)
}

type playlistUsecase struct {
	playlistRepo   repository.PlaylistRepository
	savedNoteRepo  repository.SavedNoteRepository
	noteRepo       repository.NoteRepository
	noteAccessRepo repository.NoteAccessRepository
}

func NewPlaylistUsecase(
	playlistRepo repository.PlaylistRepository,
	savedNoteRepo repository.SavedNoteRepository,
	noteRepo repository.NoteRepository,
	noteAccessRepo repository.NoteAccessRepository,
) PlaylistUsecase {
	return &playlistUsecase{
		playlistRepo:   playlistRepo,
		savedNoteRepo:  savedNoteRepo,
		noteRepo:       noteRepo,
		noteAccessRepo: noteAccessRepo,
	}
}

func validatePlaylistVisibility(v entity.Visibility) bool {
	return v == entity.VisibilityPrivate || v == entity.VisibilityPublic
}

func (u *playlistUsecase) Create(userID uint, input CreatePlaylistInput) (*entity.Playlist, error) {
	if input.Name == "" {
		return nil, ErrInvalidPlaylistName
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = entity.VisibilityPrivate
	}
	if !validatePlaylistVisibility(visibility) {
		return nil, ErrInvalidVisibility
	}

	playlist := &entity.Playlist{
		UserID:     userID,
		Name:       input.Name,
		Visibility: visibility,
	}
	if err := u.playlistRepo.Create(playlist); err != nil {
		return nil, err
	}
	return playlist, nil
}

func (u *playlistUsecase) ListMine(userID uint) ([]entity.Playlist, error) {
	return u.playlistRepo.ListByUserID(userID)
}

// checkOwnership adalah helper dipakai Update/Delete/AddNote/RemoveNote --
// semuanya SAMA-SAMA harus dilakukan pemilik playlist.
func (u *playlistUsecase) checkOwnership(userID, playlistID uint) (*entity.Playlist, error) {
	playlist, err := u.playlistRepo.FindByID(playlistID)
	if err != nil {
		return nil, err
	}
	if playlist == nil {
		return nil, ErrPlaylistNotFound
	}
	if playlist.UserID != userID {
		return nil, ErrForbidden
	}
	return playlist, nil
}

func (u *playlistUsecase) Update(userID, playlistID uint, input UpdatePlaylistInput) (*entity.Playlist, error) {
	playlist, err := u.checkOwnership(userID, playlistID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if *input.Name == "" {
			return nil, ErrInvalidPlaylistName
		}
		playlist.Name = *input.Name
	}
	if input.Visibility != nil {
		if !validatePlaylistVisibility(*input.Visibility) {
			return nil, ErrInvalidVisibility
		}
		playlist.Visibility = *input.Visibility
	}

	if err := u.playlistRepo.Update(playlist); err != nil {
		return nil, err
	}
	return playlist, nil
}

func (u *playlistUsecase) Delete(userID, playlistID uint) error {
	if _, err := u.checkOwnership(userID, playlistID); err != nil {
		return err
	}
	return u.playlistRepo.Delete(playlistID)
}

// canViewNote mereplikasi logika INTI dari AccessMiddleware (public/
// private/shared) dalam bentuk fungsi Go biasa (bukan fiber.Handler) --
// dibutuhkan di sini karena AddNote menerima note_id lewat BODY (bukan
// URL parameter), jadi tidak lewat AccessMiddleware sama sekali. Sengaja
// TIDAK menyalin ulang penuh (skip pengecekan Owner note_role dsb yang
// tidak relevan di sini) -- cuma inti "boleh lihat atau tidak".
func (u *playlistUsecase) canViewNote(userID uint, note *entity.Note) (bool, error) {
	if note.UserID == userID {
		return true, nil
	}
	switch note.Visibility {
	case entity.VisibilityPublic:
		return true, nil
	case entity.VisibilityPrivate:
		return false, nil
	case entity.VisibilityShared:
		access, err := u.noteAccessRepo.FindByNoteAndUser(note.ID, userID)
		if err != nil {
			return false, err
		}
		return access != nil, nil
	default:
		return false, nil
	}
}

// AddNote: memasukkan sebuah note ke playlist. Otomatis MEN-SAVE note
// itu dulu (lewat GetOrCreate) kalau belum pernah ke-save -- ini yang
// menjamin invariant "note di playlist manapun pasti muncul di Recently
// Saved" (sesuai keputusan desain kamu), user tidak perlu save manual
// sebagai langkah terpisah sebelum bisa masukin ke playlist.
func (u *playlistUsecase) AddNote(userID, playlistID, noteID uint) error {
	if _, err := u.checkOwnership(userID, playlistID); err != nil {
		return err
	}

	note, err := u.noteRepo.FindByID(noteID)
	if err != nil {
		return err
	}
	if note == nil {
		return ErrNoteNotFound
	}

	canView, err := u.canViewNote(userID, note)
	if err != nil {
		return err
	}
	if !canView {
		return ErrForbidden
	}

	savedNote, err := u.savedNoteRepo.GetOrCreate(userID, noteID)
	if err != nil {
		return err
	}

	return u.playlistRepo.AddNote(playlistID, savedNote.ID)
}

// RemoveNote HANYA menghapus keanggotaan note dari PLAYLIST INI --
// TIDAK meng-unsave note-nya (baris saved_notes TIDAK disentuh). Note
// tersebut tetap muncul di Recently Saved dan di playlist LAIN (kalau
// ada), cuma hilang dari playlist yang ini.
func (u *playlistUsecase) RemoveNote(userID, playlistID, noteID uint) error {
	if _, err := u.checkOwnership(userID, playlistID); err != nil {
		return err
	}

	savedNote, err := u.savedNoteRepo.FindByUserAndNote(userID, noteID)
	if err != nil {
		return err
	}
	if savedNote == nil {
		return ErrNoteNotFound
	}

	return u.playlistRepo.RemoveNote(playlistID, savedNote.ID)
}

func (u *playlistUsecase) GetNotes(requesterID *uint, playlistID uint) (*entity.Playlist, []entity.Note, error) {
	playlist, err := u.playlistRepo.FindByID(playlistID)
	if err != nil {
		return nil, nil, err
	}
	if playlist == nil {
		return nil, nil, ErrPlaylistNotFound
	}

	isOwner := requesterID != nil && *requesterID == playlist.UserID
	if playlist.Visibility == entity.VisibilityPrivate && !isOwner {
		return nil, nil, ErrForbidden
	}

	entries, err := u.playlistRepo.ListNotes(playlistID)
	if err != nil {
		return nil, nil, err
	}

	notes := make([]entity.Note, 0, len(entries))
	for _, e := range entries {
		notes = append(notes, e.SavedNote.Note)
	}
	return playlist, notes, nil
}
