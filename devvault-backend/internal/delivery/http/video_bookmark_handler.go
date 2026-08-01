package http

import (
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/pkg/youtube"
	"github.com/gofiber/fiber/v2"
)

type VideoBookmarkHandler struct {
	bookmarkUsecase usecase.VideoBookmarkUsecase
}

func NewVideoBookmarkHandler(bookmarkUsecase usecase.VideoBookmarkUsecase) *VideoBookmarkHandler {
	return &VideoBookmarkHandler{bookmarkUsecase: bookmarkUsecase}
}

type addBookmarkRequest struct {
	YoutubeURL   string `json:"youtube_url"`
	TimestampSec int    `json:"timestamp_sec"`
	NoteText     string `json:"note_text"`
}

func (h *VideoBookmarkHandler) Add(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)
	role, _ := c.Locals("note_role").(string)

	var req addBookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.YoutubeURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "youtube_url is required"})
	}

	bookmark, err := h.bookmarkUsecase.Add(note.ID, role, req.YoutubeURL, req.TimestampSec, req.NoteText)
	if err != nil {
		return mapBookmarkError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(bookmark)
}

func (h *VideoBookmarkHandler) List(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)

	bookmarks, err := h.bookmarkUsecase.List(note.ID)
	if err != nil {
		return mapBookmarkError(c, err)
	}

	return c.JSON(bookmarks)
}

func (h *VideoBookmarkHandler) Remove(c *fiber.Ctx) error {
	note := c.Locals("note").(*entity.Note)
	role, _ := c.Locals("note_role").(string)

	bookmarkID, err := strconv.ParseUint(c.Params("bookmarkId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid bookmark id"})
	}

	if err := h.bookmarkUsecase.Remove(note.ID, role, uint(bookmarkID)); err != nil {
		return mapBookmarkError(c, err)
	}

	return c.JSON(fiber.Map{"message": "bookmark deleted successfully"})
}

func mapBookmarkError(c *fiber.Ctx, err error) error {
	switch err {
	case usecase.ErrForbidden:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrBookmarkNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case youtube.ErrInvalidURL:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}
}
