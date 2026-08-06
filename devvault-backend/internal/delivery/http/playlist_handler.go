package http

import (
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type PlaylistHandler struct {
	playlistUsecase usecase.PlaylistUsecase
}

func NewPlaylistHandler(playlistUsecase usecase.PlaylistUsecase) *PlaylistHandler {
	return &PlaylistHandler{playlistUsecase: playlistUsecase}
}

type createPlaylistRequest struct {
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

func (h *PlaylistHandler) Create(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	var req createPlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	playlist, err := h.playlistUsecase.Create(userID, usecase.CreatePlaylistInput{
		Name:       req.Name,
		Visibility: entity.Visibility(req.Visibility),
	})
	if err != nil {
		return mapPlaylistError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(playlist)
}

func (h *PlaylistHandler) ListMine(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	playlists, err := h.playlistUsecase.ListMine(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch playlists"})
	}

	return c.JSON(playlists)
}

type updatePlaylistRequest struct {
	Name       *string `json:"name"`
	Visibility *string `json:"visibility"`
}

func (h *PlaylistHandler) Update(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid playlist id"})
	}

	var req updatePlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	input := usecase.UpdatePlaylistInput{Name: req.Name}
	if req.Visibility != nil {
		v := entity.Visibility(*req.Visibility)
		input.Visibility = &v
	}

	playlist, err := h.playlistUsecase.Update(userID, uint(playlistID), input)
	if err != nil {
		return mapPlaylistError(c, err)
	}

	return c.JSON(playlist)
}

func (h *PlaylistHandler) Delete(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid playlist id"})
	}

	if err := h.playlistUsecase.Delete(userID, uint(playlistID)); err != nil {
		return mapPlaylistError(c, err)
	}

	return c.JSON(fiber.Map{"message": "playlist deleted successfully"})
}

type addNoteToPlaylistRequest struct {
	NoteID uint `json:"note_id"`
}

func (h *PlaylistHandler) AddNote(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid playlist id"})
	}

	var req addNoteToPlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.NoteID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "note_id is required"})
	}

	if err := h.playlistUsecase.AddNote(userID, uint(playlistID), req.NoteID); err != nil {
		return mapPlaylistError(c, err)
	}

	return c.JSON(fiber.Map{"message": "note added to playlist"})
}

func (h *PlaylistHandler) RemoveNote(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid playlist id"})
	}
	noteID, err := strconv.ParseUint(c.Params("noteId"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid note id"})
	}

	if err := h.playlistUsecase.RemoveNote(userID, uint(playlistID), uint(noteID)); err != nil {
		return mapPlaylistError(c, err)
	}

	return c.JSON(fiber.Map{"message": "note removed from playlist"})
}

// GetNotes menangani GET /api/v1/playlists/:id -- boleh diakses TANPA
// login kalau playlist-nya public (mirip GetBySlug untuk note).
func (h *PlaylistHandler) GetNotes(c *fiber.Ctx) error {
	playlistID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid playlist id"})
	}

	var requesterID *uint
	if uidRaw := c.Locals("user_id"); uidRaw != nil {
		if uid, ok := uidRaw.(uint); ok {
			requesterID = &uid
		}
	}

	playlist, notes, err := h.playlistUsecase.GetNotes(requesterID, uint(playlistID))
	if err != nil {
		return mapPlaylistError(c, err)
	}

	return c.JSON(fiber.Map{
		"playlist": playlist,
		"notes":    notes,
	})
}

func mapPlaylistError(c *fiber.Ctx, err error) error {
	switch err {
	case usecase.ErrPlaylistNotFound, usecase.ErrNoteNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrForbidden:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrInvalidPlaylistName, usecase.ErrInvalidVisibility:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}
}
