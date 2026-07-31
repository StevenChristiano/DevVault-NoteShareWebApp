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
// ini atau tidak, dan sebagai apa?" — beda dari AuthMiddleware yang cuma
// menjawab "siapa kamu?".
//
// MODEL FINAL (Visibility dan note_access INDEPENDEN satu sama lain):
//   - Visibility menjawab "siapa boleh LIHAT baseline-nya":
//       private -> cuma Owner.
//       public  -> semua orang, termasuk guest.
//       shared  -> WAJIB terdaftar di note_accesses (whitelist, siapapun
//                  yang tidak terdaftar TIDAK BISA lihat sama sekali).
//   - note_accesses menjawab "siapa boleh EDIT" -- SELALU dicek terpisah,
//     TIDAK PEDULI visibility-nya apa. Artinya: Owner bisa kasih role
//     "editor" ke orang lain walau note-nya sedang public, dan orang itu
//     tetap bisa edit note itu meski visibility tidak pernah diubah jadi
//     shared. Mengubah visibility TIDAK PERNAH menghapus/mengubah baris
//     note_accesses yang sudah ada.
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

		// case 1: Owner can access their note and edit their own public note whatever visibility it is.
		if isLoggedIn && note.UserID == userID {
			c.Locals("note", note)
			c.Locals("note_role", "owner")
			return c.Next()
		}

		// CASE Not Owner -> check if logged in or not
		var accessRole string
		if isLoggedIn {
			access, err := accessRepo.FindByNoteAndUser(note.ID, userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to check access permissions"})
			}
			if access != nil {
				accessRole = string(access.Role)
			}
		}

		switch note.Visibility {
		case entity.VisibilityPublic:	//everyone can access public note, but only logged in user can have role (viewer/editor)
			role := "viewer"
			if accessRole == "editor" {
				role = "editor"
			}
			c.Locals("note", note)
			c.Locals("note_role", role)
			return c.Next()

		case entity.VisibilityPrivate: // only owner can access private note, but we already checked owner case above, so if not owner, return forbidden
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "this note is private"})
		
		case entity.VisibilityShared:	// only logged in user can access shared note, and must have access role (viewer/editor)
			if !isLoggedIn {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "You must be logged in to access this resource"})
			}
			if accessRole == "" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "you do not have access to this note"})
			}
			c.Locals("note", note)
			c.Locals("note_role", accessRole)
			return c.Next()

		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid note visibility"})
		}
	}
}
