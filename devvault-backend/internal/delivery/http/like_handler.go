package http

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type LikeHandler struct {
	likeUsecase usecase.LikeUsecase
}

func NewLikeHandler(likeUsecase usecase.LikeUsecase) *LikeHandler {
	return &LikeHandler{likeUsecase: likeUsecase}
}

// Toggle menangani POST /api/v1/notes/:id/like. Dipasang di belakang
// AuthMiddleware (mandatory) + AccessMiddleware -- guest tidak boleh
// like, dan note-nya harus BISA DILIHAT dulu (tidak masuk akal like note
// yang bahkan tidak boleh kamu buka).
func (h *LikeHandler) Toggle(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	note := c.Locals("note").(*entity.Note)

	liked, err := h.likeUsecase.Toggle(note.ID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to toggle like"})
	}

	return c.JSON(fiber.Map{"liked": liked})
}
