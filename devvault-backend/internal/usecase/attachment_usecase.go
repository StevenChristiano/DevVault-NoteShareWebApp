package usecase

import (
	"errors"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/fileupload"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/storage"
)

var ErrAttachmentNotFound = errors.New("attachment not found")

type AttachmentUsecase interface {
	// Upload menerima BYTE MENTAH file (bukan *multipart.FileHeader) --
	// usecase ini sengaja tidak tahu apa-apa soal Fiber/multipart, semua
	// ekstraksi bytes dari request itu tugas handler. `originalFilename`
	// dipakai untuk validasi ekstensi & disimpan ke kolom FileName.
	Upload(noteID uint, role string, originalFilename string, content []byte) (*entity.Attachment, error)
	List(noteID uint) ([]entity.Attachment, error)
	Remove(noteID uint, role string, attachmentID uint) error
}

type attachmentUsecase struct {
	attachmentRepo repository.AttachmentRepository
	storage        *storage.LocalStorage
}

func NewAttachmentUsecase(attachmentRepo repository.AttachmentRepository, storage *storage.LocalStorage) AttachmentUsecase {
	return &attachmentUsecase{attachmentRepo: attachmentRepo, storage: storage}
}

func (u *attachmentUsecase) Upload(noteID uint, role string, originalFilename string, content []byte) (*entity.Attachment, error) {
	if !canEdit(role) {
		return nil, ErrForbidden
	}

	if len(content) > fileupload.MaxFileSize {
		return nil, fileupload.ErrFileTooLarge
	}

	ext, err := fileupload.ValidateExtension(originalFilename)
	if err != nil {
		return nil, err
	}
	if err := fileupload.ValidateContent(ext, content); err != nil {
		return nil, err
	}

	storedFilename := fileupload.GenerateUniqueFilename(ext)
	if err := u.storage.Save(storedFilename, content); err != nil {
		return nil, err
	}

	attachment := &entity.Attachment{
		NoteID:   noteID,
		FileName: originalFilename,
		FilePath: storedFilename,
		FileType: detectMimeType(ext),
	}

	if err := u.attachmentRepo.Create(attachment); err != nil {
		// Rollback file fisik yang sudah kepalang disimpan, supaya tidak
		// ada "file sampah" nempel di disk tanpa ada record-nya di
		// database sama sekali kalau insert-nya gagal.
		_ = u.storage.Delete(storedFilename)
		return nil, err
	}

	return attachment, nil
}

func (u *attachmentUsecase) List(noteID uint) ([]entity.Attachment, error) {
	return u.attachmentRepo.ListByNoteID(noteID)
}

func (u *attachmentUsecase) Remove(noteID uint, role string, attachmentID uint) error {
	if !canEdit(role) {
		return ErrForbidden
	}

	attachment, err := u.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return err
	}
	if attachment == nil {
		return ErrAttachmentNotFound
	}
	if attachment.NoteID != noteID {
		return ErrAttachmentNotFound
	}

	if err := u.attachmentRepo.Delete(attachmentID); err != nil {
		return err
	}
	// File fisik dihapus SETELAH baris database berhasil dihapus -- kalau
	// urutannya dibalik dan penghapusan baris database gagal, kita akan
	// kehilangan file tanpa record-nya (lebih baik "record tanpa file"
	// yang mudah dideteksi, daripada "file tanpa record" yang jadi sampah
	// tak terlacak).
	return u.storage.Delete(attachment.FilePath)
}

// detectMimeType mengembalikan MIME type kanonik untuk disimpan ke
// kolom FileType, berdasarkan ekstensi yang SUDAH divalidasi.
func detectMimeType(ext string) string {
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}
