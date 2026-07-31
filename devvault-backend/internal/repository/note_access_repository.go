package repository

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NoteAccessRepository dipakai AccessMiddleware untuk jawab pertanyaan:
// "untuk user ini, apakah dia punya role tertentu (viewer/editor) di
// note ini?" -- dan sekarang juga dipakai NoteAccessUsecase untuk
// mengelola (tambah/ubah/hapus/lihat) baris-baris akses tersebut.
type NoteAccessRepository interface {
	FindByNoteAndUser(noteID, userID uint) (*entity.NoteAccess, error)
	
	// Upsert membuat baris baru KALAU belum ada, atau meng-update role-nya
	// KALAU sudah ada -- satu method untuk 2 skenario ("grant" & "ubah
	// role") sesuai keputusan desain di NoteAccessUsecase.
	Upsert(noteID, userID uint, role entity.Role) error
	Delete(noteID, userID uint) error

	// ListByNoteID juga meng-Preload User, supaya usecase/handler bisa
	// langsung tampilkan email/username tiap orang yang punya akses,
	// tanpa query tambahan satu-satu.
	ListByNoteID(noteID uint) ([]entity.NoteAccess, error)
}

type noteAccessRepository struct {
	db *gorm.DB
}

func NewNoteAccessRepository(db *gorm.DB) NoteAccessRepository {
	return &noteAccessRepository{db: db}
}

func (r *noteAccessRepository) FindByNoteAndUser(noteID, userID uint) (*entity.NoteAccess, error) {
	var access entity.NoteAccess
	err := r.db.Where("note_id = ? AND user_id = ?", noteID, userID).First(&access).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &access, nil
}

// Upsert pakai clause.OnConflict dari GORM: coba INSERT dulu, tapi kalau
// ternyata sudah ada baris dengan (note_id, user_id) yang sama (ingat,
// ini kombinasi unique index yang kita pasang di Tahap 1), otomatis
// UPDATE kolom `role`-nya saja -- SATU query ke database, bukan
// "SELECT dulu buat cek ada/tidak, baru INSERT atau UPDATE manual"
// (yang butuh 2 query terpisah dan rawan race condition kalau ada 2
// request Upsert nyaris bersamaan).
func (r *noteAccessRepository) Upsert(noteID, userID uint, role entity.Role) error {
	access := entity.NoteAccess{
		NoteID: noteID,
		UserID: userID,
		Role:   role,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "note_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role"}),
	}).Create(&access).Error
}

func (r *noteAccessRepository) Delete(noteID, userID uint) error {
	return r.db.Where("note_id = ? AND user_id = ?", noteID, userID).Delete(&entity.NoteAccess{}).Error
}

func (r *noteAccessRepository) ListByNoteID(noteID uint) ([]entity.NoteAccess, error) {
	var accesses []entity.NoteAccess
	err := r.db.Preload("User").Where("note_id = ?", noteID).Find(&accesses).Error
	if err != nil {
		return nil, err
	}
	return accesses, nil
}
