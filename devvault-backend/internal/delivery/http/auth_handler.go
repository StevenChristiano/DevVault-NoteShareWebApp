// Package http berisi HANDLER (controller) Fiber + routing.
// Handler menerima request HTTP mentah, terjemahkan jadi pemanggilan
// usecase, lalu terjemahkan hasilnya balik jadi response JSON. Handler
// TIDAK BOLEH berisi logika bisnis (validasi aturan, hashing, dst) —
// itu semua sudah ada di usecase, di sini cuma "penerjemah".
package http

import (
	"time"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request body is not valid"})
	}

	if req.Username == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username and email are required",
		})
	}

	if err := h.authUsecase.Register(req.Username, req.Email, req.Password); err != nil {
		switch err {
    case usecase.ErrEmailAlreadyUsed:
        return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
    case usecase.ErrPasswordTooWeak:
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    default:
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to register user"})
    }
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "registration successful"})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "request body is not valid"})
	}

	jwtToken, err := h.authUsecase.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "email or password is incorrect"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    jwtToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
		// Secure: true, // WAJIB diaktifkan kalau sudah deploy pakai HTTPS.
		// Dibiarkan false dulu karena development masih pakai http://localhost.
	})

	return c.JSON(fiber.Map{"message": "login successful"})
}
