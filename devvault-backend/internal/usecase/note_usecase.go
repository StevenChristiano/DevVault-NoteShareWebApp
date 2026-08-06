package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/slug"
)

var (
	ErrNoteNotFound  = errors.New("note not found")
	ErrForbidden     = errors.New("you don't have permission to perform this action")
	ErrInvalidInput  = errors.New("header is required")
)

// reservedSlugs adalah kata-kata yang TIDAK BOLEH jadi slug note, karena
// sudah dipakai sebagai path STATIS di routing (mis. "/notes/saved").
// Kalau slug note kebetulan sama dengan salah satu ini, note tersebut
// akan "tidak terjangkau" lewat GET /notes/:slug (selalu ke-intercept
// duluan sama route statis-nya). Dicek di generateUniqueSlug -- MEKANISME
// SAMA dengan penanganan slug kembar (bukan pengecualian terpisah),
// karena secara konsep dua-duanya masalah yang identik: "slug ini tidak
// boleh dipakai, cari alternatif lain".
var reservedSlugs = map[string]bool{
	"saved": true,
}

type CreateNoteInput struct {
	Header     string
	Paragraph  string
	Visibility entity.Visibility
}

type UpdateNoteInput struct {
	Header     *string
	Paragraph  *string
	Visibility *entity.Visibility
}

type NoteUsecase interface {
	Create(userID uint, input CreateNoteInput) (*entity.Note, error)
	GetBySlugForView(note *entity.Note) (*entity.Note, error)
	Update(userID uint, noteID uint, role string, input UpdateNoteInput) (*entity.Note, error)
	Delete(userID uint, noteID uint, role string) error
	ListMyNotes(userID uint) ([]entity.Note, error)
}

type noteUsecase struct {
	noteRepo repository.NoteRepository
}

func NewNoteUsecase(noteRepo repository.NoteRepository) NoteUsecase {
	return &noteUsecase{noteRepo: noteRepo}
}

func (u *noteUsecase) Create(userID uint, input CreateNoteInput) (*entity.Note, error) {
	if input.Header == "" {
		return nil, ErrInvalidInput
	}

	visibility := input.Visibility
	if visibility == "" {
		visibility = entity.VisibilityPrivate
	}

	note := &entity.Note{
		Header:     input.Header,
		Paragraph:  input.Paragraph,
		UserID:     userID,
		Visibility: visibility,
	}

	if visibility == entity.VisibilityPublic {
		generatedSlug, err := u.generateUniqueSlug(input.Header)
		if err != nil {
			return nil, err
		}
		note.Slug = generatedSlug
	}

	if err := u.noteRepo.Create(note); err != nil {
		return nil, err
	}
	return note, nil
}

func (u *noteUsecase) generateUniqueSlug(header string) (string, error) {
	base := slug.Generate(header)
	candidate := base

	for i := 0; i < 5; i++ {
		if reservedSlugs[candidate] {
			candidate = base + "-" + slug.RandomSuffix()
			continue
		}
		exists, err := u.noteRepo.ExistsBySlug(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = base + "-" + slug.RandomSuffix()
	}

	return candidate, nil
}

// GetBySlugForView dipanggil handler SETELAH AccessMiddleware sudah
// memastikan user boleh lihat note ini (note diambil dari c.Locals).
// Method ini HANYA menambah efek samping "tambah view_count" — bukan
// mengambil ulang data note (itu sudah dilakukan middleware, tidak perlu
// query dua kali).
func (u *noteUsecase) GetBySlugForView(note *entity.Note) (*entity.Note, error) {
	if err := u.noteRepo.IncrementViewCount(note.ID); err != nil {
		return nil, err
	}
	// Refleksikan penambahan itu ke objek yang dikembalikan ke handler,
	// supaya response yang dikirim ke client sudah menampilkan angka
	// terbaru tanpa perlu query ulang ke database.
	note.ViewCount++
	return note, nil
}

// Update menerapkan aturan role dari dokumen:
//   - owner  -> boleh ubah semua field, termasuk Visibility.
//   - editor -> HANYA boleh ubah Header/Paragraph, TIDAK BOLEH ubah Visibility.
//   - selain itu (viewer, atau role tidak dikenali) -> ditolak sama sekali.
func (u *noteUsecase) Update(userID uint, noteID uint, role string, input UpdateNoteInput) (*entity.Note, error) {
	note, err := u.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}

	if role != "owner" && role != "editor" {
		return nil, ErrForbidden
	}

	if input.Header != nil {
		note.Header = *input.Header
	}
	if input.Paragraph != nil {
		note.Paragraph = *input.Paragraph
	}

	if input.Visibility != nil {
		if role != "owner" {
			return nil, ErrForbidden
		}
		note.Visibility = *input.Visibility

		// Kalau diubah JADI public dan belum punya slug (mis. sebelumnya
		// private), generate slug baru sekarang.
		if note.Visibility == entity.VisibilityPublic && note.Slug == "" {
			generatedSlug, err := u.generateUniqueSlug(note.Header)
			if err != nil {
				return nil, err
			}
			note.Slug = generatedSlug
		}
	}

	if err := u.noteRepo.Update(note); err != nil {
		return nil, err
	}
	return note, nil
}

// Delete: HANYA owner yang boleh, sesuai tabel role di dokumen spek.
func (u *noteUsecase) Delete(userID uint, noteID uint, role string) error {
	if role != "owner" {
		return ErrForbidden
	}

	note, err := u.noteRepo.FindByID(noteID)
	if err != nil {
		return err
	}
	if note == nil {
		return ErrNoteNotFound
	}

	return u.noteRepo.Delete(noteID)
}

func (u *noteUsecase) ListMyNotes(userID uint) ([]entity.Note, error) {
	return u.noteRepo.FindByUserID(userID)
}
