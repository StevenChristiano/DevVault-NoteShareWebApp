package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
)

var (
	ErrUserNotFound      = errors.New("user with that email was not found")
	ErrInvalidRole       = errors.New("role must be either 'viewer' or 'editor'")
	ErrCannotGrantToSelf = errors.New("you cannot grant access to yourself")
)

// AccessInfo adalah bentuk data yang dikembalikan ke handler untuk
// ditampilkan ke client -- SENGAJA bukan langsung entity.NoteAccess,
// supaya kita bisa "ratakan" data User (ambil cuma Username/Email-nya,
// tanpa ikut nge-expose PasswordHash dkk) tanpa perlu json:"-" tag di
// entity.User yang bisa bikin bingung dipakai di tempat lain.
type AccessInfo struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type NoteAccessUsecase interface {
	// Grant menambah ATAU mengubah role user (dicari lewat email) untuk
	// sebuah note. role dan visibility TIDAK saling mempengaruhi -- lihat
	// komentar di dalam method.
	Grant(requesterID, noteID uint, targetEmail string, role string) error
	Revoke(requesterID, noteID uint, targetEmail string) error
	List(requesterID, noteID uint) ([]AccessInfo, error)
}

type noteAccessUsecase struct {
	noteRepo   repository.NoteRepository
	userRepo   repository.UserRepository
	accessRepo repository.NoteAccessRepository
}

func NewNoteAccessUsecase(
	noteRepo repository.NoteRepository,
	userRepo repository.UserRepository,
	accessRepo repository.NoteAccessRepository,
) NoteAccessUsecase {
	return &noteAccessUsecase{
		noteRepo:   noteRepo,
		userRepo:   userRepo,
		accessRepo: accessRepo,
	}
}

// checkIsOwner adalah helper yang dipakai Grant/Revoke/List -- ketiganya
// SAMA-SAMA harus dilakukan Owner Only (sesuai API contract di dokumen).
// Sengaja query ulang note-nya di sini (bukan cuma percaya role dari
// AccessMiddleware), supaya usecase ini tetap AMAN dipanggil dari jalur
// manapun (mis. nanti dari test, atau dari usecase lain), tidak
// bergantung penuh pada middleware HTTP yang notabene bisa saja beda
// setup-nya di masa depan.
func (u *noteAccessUsecase) checkIsOwner(requesterID, noteID uint) (*entity.Note, error) {
	note, err := u.noteRepo.FindByID(noteID)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, ErrNoteNotFound
	}
	if note.UserID != requesterID {
		return nil, ErrForbidden
	}
	return note, nil
}

// Grant memberi ATAU mengubah role (viewer/editor) untuk user tertentu
// (dicari lewat email) pada sebuah note.
//
// KEPUTUSAN DESAIN PENTING: Grant TIDAK PERNAH mengubah note.Visibility.
// Visibility dan role akses itu independen (sesuai keputusan kamu) --
// Owner bisa kasih akses editor ke Budi walau note-nya sedang `public`
// ATAU `private` sekalipun (walau kalau private, akses itu jadi "tidak
// berpengaruh" karena private tetap mutlak cuma-Owner di level VIEW --
// lihat AccessMiddleware). Ini konsisten: Grant selalu boleh dipanggil,
// efeknya yang baru benar-benar "terasa" tergantung visibility saat itu.
func (u *noteAccessUsecase) Grant(requesterID, noteID uint, targetEmail string, role string) error {
	if _, err := u.checkIsOwner(requesterID, noteID); err != nil {
		return err
	}

	parsedRole := entity.Role(role)
	if parsedRole != entity.RoleViewer && parsedRole != entity.RoleEditor {
		return ErrInvalidRole
	}

	targetUser, err := u.userRepo.FindByEmail(targetEmail)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return ErrUserNotFound
	}
	if targetUser.ID == requesterID {
		return ErrCannotGrantToSelf
	}

	return u.accessRepo.Upsert(noteID, targetUser.ID, parsedRole)
}

func (u *noteAccessUsecase) Revoke(requesterID, noteID uint, targetEmail string) error {
	if _, err := u.checkIsOwner(requesterID, noteID); err != nil {
		return err
	}

	targetUser, err := u.userRepo.FindByEmail(targetEmail)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return ErrUserNotFound
	}

	return u.accessRepo.Delete(noteID, targetUser.ID)
}

func (u *noteAccessUsecase) List(requesterID, noteID uint) ([]AccessInfo, error) {
	if _, err := u.checkIsOwner(requesterID, noteID); err != nil {
		return nil, err
	}

	accesses, err := u.accessRepo.ListByNoteID(noteID)
	if err != nil {
		return nil, err
	}

	result := make([]AccessInfo, 0, len(accesses))
	for _, a := range accesses {
		result = append(result, AccessInfo{
			UserID:   a.UserID,
			Username: a.User.Username,
			Email:    a.User.Email,
			Role:     string(a.Role),
		})
	}
	return result, nil
}
