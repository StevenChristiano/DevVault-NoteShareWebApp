package http

import (
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type FeedHandler struct {
	feedUsecase usecase.FeedUsecase
}

func NewFeedHandler(feedUsecase usecase.FeedUsecase) *FeedHandler {
	return &FeedHandler{feedUsecase: feedUsecase}
}

// GetFYP menangani GET /api/v1/fyp -- Public (boleh diakses tanpa
// login), makanya dipasang dengan OptionalAuthMiddleware di routes.go.
// Query params: ?sort=likes|saves|latest (default latest),
// &following=true (cuma berlaku kalau sedang login), &page=1.
func (h *FeedHandler) GetFYP(c *fiber.Ctx) error {
	sortBy := c.Query("sort", "latest")
	followingOnly := c.Query("following") == "true"

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	var userID *uint
	if uidRaw := c.Locals("user_id"); uidRaw != nil {
		if uid, ok := uidRaw.(uint); ok {
			userID = &uid
		}
	}

	notes, err := h.feedUsecase.GetFYP(sortBy, userID, followingOnly, page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load feed"})
	}

	return c.JSON(notes)
}
