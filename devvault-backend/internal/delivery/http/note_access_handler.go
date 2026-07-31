package http

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type NoteAccessHandler struct {
	accessUsecase usecase.NoteAccessUsecase
}

func NewNoteAccessHandler(accessUsecase usecase.NoteAccessUsecase) *NoteAccessHandler {
	return &NoteAccessHandler{accessUsecase: accessUsecase}
}

type grantAccessRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Grant menangani POST /api/v1/notes/:id/access -- "Memberi/mengubah
// akses" sesuai API contract (satu endpoint, upsert).
func (h *NoteAccessHandler) Grant(c *fiber.Ctx) error {
	requesterID := c.Locals("user_id").(uint)
	note := c.Locals("note").(*entity.Note)

	var req grantAccessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	if err := h.accessUsecase.Grant(requesterID, note.ID, req.Email, req.Role); err != nil {
		return mapAccessError(c, err)
	}

	return c.JSON(fiber.Map{"message": "access granted successfully"})
}

type revokeAccessRequest struct {
	Email string `json:"email"`
}

// Revoke menangani DELETE /api/v1/notes/:id/access. Body dipakai
// (bukan path parameter) karena yang mau dihapus itu SATU baris
// spesifik (kombinasi note+user), dan client paling gampang menyediakan
// email (bukan mencari tahu user_id-nya sendiri).
func (h *NoteAccessHandler) Revoke(c *fiber.Ctx) error {
	requesterID := c.Locals("user_id").(uint)
	note := c.Locals("note").(*entity.Note)

	var req revokeAccessRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email is required"})
	}

	if err := h.accessUsecase.Revoke(requesterID, note.ID, req.Email); err != nil {
		return mapAccessError(c, err)
	}

	return c.JSON(fiber.Map{"message": "access revoked successfully"})
}

// List menangani GET /api/v1/notes/:id/access -- Owner melihat daftar
// siapa saja yang punya akses ke note-nya.
func (h *NoteAccessHandler) List(c *fiber.Ctx) error {
	requesterID := c.Locals("user_id").(uint)
	note := c.Locals("note").(*entity.Note)

	accesses, err := h.accessUsecase.List(requesterID, note.ID)
	if err != nil {
		return mapAccessError(c, err)
	}

	return c.JSON(accesses)
}

// mapAccessError menerjemahkan error dari usecase ke status HTTP yang
// tepat. Dipisah jadi fungsi sendiri karena 3 handler di atas (Grant,
// Revoke, List) semuanya butuh pemetaan error yang SAMA PERSIS --
// menghindari copy-paste switch-case yang sama 3 kali.
func mapAccessError(c *fiber.Ctx, err error) error {
	switch err {
	case usecase.ErrNoteNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrForbidden:
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrUserNotFound:
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrInvalidRole:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	case usecase.ErrCannotGrantToSelf:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "something went wrong"})
	}
}
