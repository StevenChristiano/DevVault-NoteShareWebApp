package http

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type SaveHandler struct {
	saveUsecase usecase.SaveUsecase
}

func NewSaveHandler(saveUsecase usecase.SaveUsecase) *SaveHandler {
	return &SaveHandler{saveUsecase: saveUsecase}
}

func (h *SaveHandler) Toggle(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	note := c.Locals("note").(*entity.Note)

	saved, err := h.saveUsecase.Toggle(note.ID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to toggle save"})
	}

	return c.JSON(fiber.Map{"saved": saved})
}

// ListSaved menangani GET /api/v1/notes/saved -- TIDAK butuh
// AccessMiddleware (bukan soal 1 note spesifik), cukup AuthMiddleware
// biasa, sama pola dengan NoteHandler.ListMine.
func (h *SaveHandler) ListSaved(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	notes, err := h.saveUsecase.ListSaved(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch saved notes"})
	}

	return c.JSON(notes)
}
