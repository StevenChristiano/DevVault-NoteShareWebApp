package middleware

import (
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// AccessMiddleware menjawab pertanyaan "kamu BOLEH akses note spesifik
// ini atau tidak?" — beda dari AuthMiddleware yang cuma menjawab "siapa
// kamu?". Middleware ini yang menerapkan aturan dari dokumen spek:
//
//   - public  -> siapa saja boleh baca, termasuk yang belum login.
//   - private -> HANYA Owner note itu sendiri.
//   - shared  -> Owner, ATAU siapa pun yang punya baris di note_accesses
//     (dengan role viewer/editor).
//
// CATATAN: middleware ini baru benar-benar "dipasang" ke sebuah route
// mulai Tahap 3 (waktu endpoint GET /api/v1/notes/:slug dibuat). Di
// Tahap 2 ini kita bangun dulu logikanya karena butuh NoteRepository &
// NoteAccessRepository yang sudah kita siapkan.
//
// PENTING: karena note bisa diakses TANPA login (kasus public), middleware
// ini HARUS dipasang SETELAH AuthMiddleware yang bersifat opsional, atau
// berdiri sendiri dan mengecek c.Locals("user_id") yang BOLEH kosong.
// Implementasi di bawah mengasumsikan yang kedua: AuthMiddleware belum
// tentu dipasang duluan, jadi user_id boleh nil (artinya guest/belum login).
func AccessMiddleware(noteRepo repository.NoteRepository, accessRepo repository.NoteAccessRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		slug := c.Params("slug")

		note, err := noteRepo.FindBySlug(slug)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch note"})
		}
		if note == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "note not found"})
		}

		// case 1: public -> everyone can access, no login required. Save note to locals and continue.
		if note.Visibility == entity.VisibilityPublic {
			c.Locals("note", note)
			return c.Next()
		}

		userIDRaw := c.Locals("user_id")
		if userIDRaw == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "you must be logged in to access this resource"})
		}
		userID, ok := userIDRaw.(uint)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid session"})
		}

		// case 2: Owner can access either view or edit.
		if note.UserID == userID {
			c.Locals("note", note)
			c.Locals("note_role", "owner")
			return c.Next()
		}

		// case 3: private but NOT owner -> reject (not shown)
		// other exception for private notes: if the note is private and the user is not the owner, they cannot access it.
		if note.Visibility == entity.VisibilityPrivate {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "This note is private"})
		}

		// case 4: shared, not owner -> check note_accesses table.
		access, err := accessRepo.FindByNoteAndUser(note.ID, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check access permissions"})
		}
		if access == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you do not have access to this note"})
		}

		c.Locals("note", note)
		c.Locals("note_role", string(access.Role))
		return c.Next()
	}
}
