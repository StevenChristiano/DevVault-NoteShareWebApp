package http

import (
	"fmt"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type NoteHandler struct {
	noteUsecase usecase.NoteUsecase
}

func NewNoteHandler(noteUsecase usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{noteUsecase: noteUsecase}
}

type createNoteRequest struct {
	Header     string `json:"header"`
	Paragraph  string `json:"paragraph"`
	Visibility string `json:"visibility"`
}

func (h *NoteHandler) Create(c *fiber.Ctx) error {
	// user_id di sini DIJAMIN ada, karena route ini dipasang di belakang
	// AuthMiddleware yang mandatory (lihat routes.go) — beda dengan
	// GetBySlug yang pakai OptionalAuthMiddleware.
	userID := c.Locals("user_id").(uint)

	var req createNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	note, err := h.noteUsecase.Create(userID, usecase.CreateNoteInput{
		Header:     req.Header,
		Paragraph:  req.Paragraph,
		Visibility: entity.Visibility(req.Visibility),
	})
	if err != nil {
		if err == usecase.ErrInvalidInput {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create note"})
	}

	return c.Status(fiber.StatusCreated).JSON(note)
}

// GetBySlug TIDAK memanggil noteRepo/FindBySlug lagi — note-nya sudah
// diambil oleh AccessMiddleware dan dititipkan lewat c.Locals("note").
// Handler ini tinggal ambil titipan itu, lalu minta usecase menambah
// view_count sebagai efek samping.
func (h *NoteHandler) GetBySlug(c *fiber.Ctx) error {
	fmt.Println("Handler slug:", c.Params("slug"))
	note := c.Locals("note").(*entity.Note)

	updated, err := h.noteUsecase.GetBySlugForView(note)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load note"})
	}

	return c.JSON(updated)
}

type updateNoteRequest struct {
	Header     *string `json:"header"`
	Paragraph  *string `json:"paragraph"`
	Visibility *string `json:"visibility"`
}

func (h *NoteHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	// note_role dititipkan AccessMiddleware ("owner", "editor", atau "viewer").
	role, _ := c.Locals("note_role").(string)
	note := c.Locals("note").(*entity.Note)

	var req updateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	input := usecase.UpdateNoteInput{
		Header:    req.Header,
		Paragraph: req.Paragraph,
	}
	if req.Visibility != nil {
		v := entity.Visibility(*req.Visibility)
		input.Visibility = &v
	}

	updated, err := h.noteUsecase.Update(userID, note.ID, role, input)
	if err != nil {
		switch err {
		case usecase.ErrNoteNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case usecase.ErrForbidden:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update note"})
		}
	}

	return c.JSON(updated)
}

func (h *NoteHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	role, _ := c.Locals("note_role").(string)
	note := c.Locals("note").(*entity.Note)

	if err := h.noteUsecase.Delete(userID, note.ID, role); err != nil {
		switch err {
		case usecase.ErrNoteNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		case usecase.ErrForbidden:
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete note"})
		}
	}

	return c.JSON(fiber.Map{"message": "note deleted successfully"})
}

func (h *NoteHandler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	notes, err := h.noteUsecase.ListMyNotes(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch notes"})
	}

	return c.JSON(notes)
}
