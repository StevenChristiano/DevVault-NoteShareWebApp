package http

import (
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type FollowHandler struct {
	followUsecase usecase.FollowUsecase
}

func NewFollowHandler(followUsecase usecase.FollowUsecase) *FollowHandler {
	return &FollowHandler{followUsecase: followUsecase}
}

// Toggle menangani POST /api/v1/users/:id/follow -- beda dari
// Like/Save, ini TIDAK melalui AccessMiddleware sama sekali (targetnya
// USER, bukan note), cukup AuthMiddleware biasa. :id di sini diambil
// LANGSUNG dari c.Params (bukan lewat resolveNote), karena tidak ada
// urusan dengan tabel notes sama sekali.
func (h *FollowHandler) Toggle(c *fiber.Ctx) error {
	followerID := c.Locals("user_id").(uint)

	targetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	following, err := h.followUsecase.Toggle(followerID, uint(targetID))
	if err != nil {
		if err == usecase.ErrCannotFollowSelf {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to toggle follow"})
	}

	return c.JSON(fiber.Map{"following": following})
}
