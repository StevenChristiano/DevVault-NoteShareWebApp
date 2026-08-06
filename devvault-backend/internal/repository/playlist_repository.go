package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlaylistRepository interface {
	Create(playlist *entity.Playlist) error
	FindByID(id uint) (*entity.Playlist, error)
	Update(playlist *entity.Playlist) error
	Delete(id uint) error
	ListByUserID(userID uint) ([]entity.Playlist, error)

	// AddNote pakai Upsert dengan DoNothing -- kalau note itu SUDAH ada
	// di playlist ini, tidak ada apa pun yang perlu diubah, jadi kalau
	// konflik (sudah ada), diamkan saja, jangan error "duplicate".
	AddNote(playlistID, savedNoteID uint) error
	RemoveNote(playlistID, savedNoteID uint) error
	// ListNotes meng-Preload sampai 2 tingkat (SavedNote, lalu Note di
	// dalam SavedNote), supaya usecase langsung dapat data Note lengkap.
	ListNotes(playlistID uint) ([]entity.PlaylistNote, error)
}

type playlistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) PlaylistRepository {
	return &playlistRepository{db: db}
}

func (r *playlistRepository) Create(playlist *entity.Playlist) error {
	return r.db.Create(playlist).Error
}

func (r *playlistRepository) FindByID(id uint) (*entity.Playlist, error) {
	var playlist entity.Playlist
	err := r.db.First(&playlist, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &playlist, nil
}

func (r *playlistRepository) Update(playlist *entity.Playlist) error {
	return r.db.Save(playlist).Error
}

// Delete cukup hapus baris playlists -- semua baris playlist_notes
// terkait ikut terhapus OTOMATIS lewat ON DELETE CASCADE yang sudah
// dipasang di entity.PlaylistNote. Note ASLI dan baris saved_notes-nya
// TIDAK TERSENTUH SAMA SEKALI (cuma "keanggotaan di playlist ini" yang
// hilang, note-nya tetap ada di Recently Saved).
func (r *playlistRepository) Delete(id uint) error {
	return r.db.Delete(&entity.Playlist{}, id).Error
}

func (r *playlistRepository) ListByUserID(userID uint) ([]entity.Playlist, error) {
	var playlists []entity.Playlist
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&playlists).Error
	if err != nil {
		return nil, err
	}
	return playlists, nil
}

func (r *playlistRepository) AddNote(playlistID, savedNoteID uint) error {
	entry := entity.PlaylistNote{PlaylistID: playlistID, SavedNoteID: savedNoteID}
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&entry).Error
}

func (r *playlistRepository) RemoveNote(playlistID, savedNoteID uint) error {
	return r.db.Where("playlist_id = ? AND saved_note_id = ?", playlistID, savedNoteID).
		Delete(&entity.PlaylistNote{}).Error
}

func (r *playlistRepository) ListNotes(playlistID uint) ([]entity.PlaylistNote, error) {
	var entries []entity.PlaylistNote
	err := r.db.Preload("SavedNote.Note").
		Where("playlist_id = ?", playlistID).
		Order("created_at DESC").
		Find(&entries).Error
	if err != nil {
		return nil, err
	}
	return entries, nil
}
