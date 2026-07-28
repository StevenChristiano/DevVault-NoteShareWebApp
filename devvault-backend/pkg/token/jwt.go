// Package token membungkus pembuatan & verifikasi JWT.
// Dipisah ke pkg/ (bukan internal/) karena ini utilitas generic — tidak
// tahu apa-apa soal User/Note/bisnis DevVault, cuma tahu "bikin token dari
// angka ID" dan "baca angka ID dari token". Kalau nanti project lain butuh
// hal serupa, package ini bisa dipakai ulang apa adanya.
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims adalah "isi" JWT yang kita titipkan.
// jwt.RegisteredClaims sudah menyediakan field standar (ExpiresAt,
// IssuedAt, dst) sesuai spesifikasi resmi JWT (RFC 7519) — kita tinggal
// nge-embed itu dan menambah UserID sebagai custom claim milik kita sendiri.
type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate membuat JWT baru untuk userID tertentu, valid selama `ttl`
// (time-to-live) sejak sekarang.
//
// SigningMethodHS256 artinya tanda tangan pakai 1 secret key yang sama
// dipakai untuk MENANDATANGANI dan MEMVERIFIKASI (symmetric). Ini cukup
// untuk project kita karena yang menandatangani & memverifikasi sama-sama
// backend kita sendiri. (Ada juga metode asymmetric/RS256 yang dipakai
// kalau pihak yang verifikasi beda server dari yang menandatangani —
// tidak relevan untuk kasus kita sekarang.)
func Generate(userID uint, secret string, ttl time.Duration) (string, error) {
	claims := Claims {
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
	
	rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return rawToken.SignedString([]byte(secret))
}

// Parse memverifikasi signature token pakai `secret`, DAN mengecek apakah
// token sudah kadaluarsa (`exp`) — dua-duanya ditangani otomatis oleh
// library jwt-go, kita tidak perlu cek manual.
func Parse(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil { // if err is not nil, it means the token is either malformed, signature is invalid, or expired
		return nil, err
	}
	if !parsedToken.Valid { // Token is not valid anymore (signature mismatch or expired)
		return nil, jwt.ErrTokenSignatureInvalid
	}

	return claims, nil
}