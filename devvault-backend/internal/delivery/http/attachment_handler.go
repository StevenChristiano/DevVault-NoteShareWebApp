package http

import (
	"io"
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/fileupload"
	"github.com/gofiber/fiber/v2"
)

type AttachmentHandler struct {
	attachmentUsecase usecase.AttachmentUsecase
}

func NewAttachmentHandler(attachmentUsecase usecase.AttachmentUsecase) *AttachmentHandler {
	return &AttachmentHandler{attachmentUsecase: attachmentUsecase}
}

// Upload menangani POST /api/v1/notes/:id/upload (multipart/form-data,
// field name "file" -- INI BEDA dari handler lain yang pakai c.BodyParser
// untuk JSON; upload file WAJIB pakai c.FormFile, bukan BodyParser).
func (h *AttachmentHandler) Upload(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)
	role, _ := c.Locals("note_role").(string)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required (field name: 'file')"})
	}

	// Buka file yang di-upload untuk dibaca isinya jadi []byte -- ini
	// dilakukan DI HANDLER (bukan usecase), karena cuma handler yang
	// tahu cara "membuka" *multipart.FileHeader milik Fiber. Usecase
	// cukup terima []byte polos, tidak perlu tahu soal Fiber sama sekali.
	src, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open uploaded file"})
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to read uploaded file"})
	}

	attachment, err := h.attachmentUsecase.Upload(note.ID, role, fileHeader.Filename, content)
	if err != nil {
		return mapAttachmentError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(attachment)
}

func (h *AttachmentHandler) List(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)

	attachments, err := h.attachmentUsecase.List(note.ID)
	if err != nil {
		return mapAttachmentError(c, err)
	}

	return c.JSON(attachments)
}

func (h *AttachmentHandler) Remove(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)
	role, _ := c.Locals("note_role").(string)

	attachmentID, err := strconv.ParseUint(c.Params("attachmentId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid attachment id"})
	}

	if err := h.attachmentUsecase.Remove(note.ID, role, uint(attachmentID)); err != nil {
		return mapAttachmentError(c, err)
	}

	return c.JSON(fiber.Map{"message": "attachment deleted successfully"})
}

func mapAttachmentError(c *fiber.Ctx, err error) error {
	switch err {
	case usecase.ErrForbidden:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrAttachmentNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case fileupload.ErrFileTooLarge, fileupload.ErrExtensionNotAllowed, fileupload.ErrMimeTypeMismatch:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}
}
