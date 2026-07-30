package middleware

import (
	// "fmt"
	"strconv"

	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/entity"
	"github.com/StevenChristiano/DevVault-NoteShareWebApp/devvault-backend/internal/repository"
	"github.com/gofiber/fiber/v2"
)

// resolveNote menemukan note dari parameter URL — mendukung DUA bentuk
// route sekaligus: "/notes/:slug" (buat endpoint publik "lihat note")
// dan "/notes/:id" (buat endpoint yang Owner-nya sendiri sudah tahu ID
// note-nya, dipakai Update/Delete). Cukup satu middleware ini yang
// dipasang ke keduanya, tidak perlu duplikasi logic akses 2 kali.
func resolveNote(c *fiber.Ctx, noteRepo repository.NoteRepository) (*entity.Note, error) {
	if idParam := c.Params("id"); idParam != "" {
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return nil, nil // id tidak valid -> diperlakukan sama seperti "tidak ketemu"
		}
		return noteRepo.FindByID(uint(id))
	}
	return noteRepo.FindBySlug(c.Params("slug"))
}

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
		// fmt.Println("OriginalURL :", c.OriginalURL())
		// fmt.Println("Route Path  :", c.Route().Path)
		// fmt.Println("Slug        :", c.Params("slug"))
		// fmt.Println("ID          :", c.Params("id"))
		note, err := resolveNote(c, noteRepo)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch note"})
		}
		if note == nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "note was not found"})
		}

		var userID uint
		isLoggedIn := false
		if userIDRaw := c.Locals("user_id"); userIDRaw != nil {
			uid, ok := userIDRaw.(uint)
			if !ok {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "sesi tidak valid"})
			}
			userID = uid
			isLoggedIn = true
		}

		// case 1: public -> REMINDER: Owner can still edit their own public note
		if isLoggedIn && note.UserID == userID {
			c.Locals("note", note)
			c.Locals("note_role", "owner")
			return c.Next()
		}

		//case 2: public -> anyone can view
		if note.Visibility == entity.VisibilityPublic {
			c.Locals("note", note)
			c.Locals("note_role", "viewer")
			return c.Next()
		}

		//case beside public (private or shared)
		if !isLoggedIn {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "You must be logged in to access this resource"})
		}

		// case 3: private -> only owner can view
		if note.Visibility == entity.VisibilityPrivate {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "this note is private"})
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
