// Package slug mengubah teks biasa jadi format URL-friendly.
// SENGAJA "bodoh" — cuma transformasi teks, tidak tahu apa-apa soal
// database atau "apakah slug ini sudah dipakai note lain". Urusan
// keunikan itu tanggung jawab usecase (lihat NoteUsecase.generateUniqueSlug),
// karena itu butuh akses ke repository buat cek ke database.
package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9\s-]`)
	multipleSpaces  = regexp.MustCompile(`[\s-]+`)
)

// Generate mengubah teks jadi slug dasar.
// Contoh: "Cara Belajar Golang!" -> "cara-belajar-golang"
func Generate(text string) string {
	s := strings.ToLower(text)
	s = nonAlphanumeric.ReplaceAllString(s, "")   // buang karakter selain huruf/angka/spasi/strip
	s = multipleSpaces.ReplaceAllString(s, "-")   // spasi (atau strip berulang) jadi satu "-"
	s = strings.Trim(s, "-")                      // buang "-" nyempil di awal/akhir

	if s == "" { // kalo hasilnya kosong, kasih default biar ga error di URL
		s = "note"
	}
	return s
}

// RandomSuffix menghasilkan string acak pendek (mis. "a3f9c2"), dipakai
// usecase untuk menabrakkan slug yang sudah kepakai jadi unik.
func RandomSuffix() string {
	b := make([]byte, 3) // 3 byte -> 6 karakter hex
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
