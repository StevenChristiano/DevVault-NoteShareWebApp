// Package fileupload berisi helper murni untuk memvalidasi & menamai
// ulang file upload -- TIDAK menyimpan file (itu tugas repository/
// usecase yang memanggil helper ini), cuma memutuskan "file ini boleh
// diterima atau tidak" dan "nama fisiknya harus apa".
package fileupload

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const MaxFileSize = 10 * 1024 * 1024 // 10 MB, sesuai dokumen spek

var (
	ErrFileTooLarge     = errors.New("file size exceeds the 10MB limit")
	ErrExtensionNotAllowed = errors.New("file extension is not allowed")
	ErrMimeTypeMismatch = errors.New("file content does not match its extension")
)

var allowedExtensions = map[string][]string{
	".pdf":  {"application/pdf"},
	".txt":  {"text/plain; charset=utf-8", "text/plain"},
	".png":  {"image/png"},
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".gif":  {"image/gif"},
	".docx": {"application/zip", "application/octet-stream"},
	".xlsx": {"application/zip", "application/octet-stream"},
}

func ValidateExtension(filename string) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedExtensions[ext]; !ok {
		return "", ErrExtensionNotAllowed
	}
	return ext, nil
}

func ValidateContent(ext string, content []byte) error {
	allowedMimes, ok := allowedExtensions[ext]
	if !ok {
		return ErrExtensionNotAllowed
	}

	sample := content
	if len(sample) > 512 {
		sample = sample[:512]
	}
	detectedMime := http.DetectContentType(sample)

	for _, allowed := range allowedMimes {
		if detectedMime == allowed {
			return nil
		}
	}
	return ErrMimeTypeMismatch
}

func GenerateUniqueFilename(ext string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), hex.EncodeToString(b), ext)
}
