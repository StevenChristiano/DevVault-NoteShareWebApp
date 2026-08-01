// Package storage membungkus operasi baca/tulis file ke disk lokal.
// Dipisah dari usecase supaya kalau nanti project ini pindah ke cloud
// storage (S3/Cloudinary, seperti disebut opsional di tech stack), yang
// perlu diganti CUMA isi package ini -- usecase (video_bookmark_usecase,
// attachment_usecase) tidak perlu tahu file-nya disimpan di mana secara
// fisik, cukup panggil Save()/Delete().
package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorage menyimpan file di folder lokal (default "./uploads").
type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

// Save menulis `content` ke file bernama `filename` di dalam baseDir.
// MkdirAll dipanggil setiap kali (bukan cuma sekali di awal) karena ini
// operasi murah kalau folder-nya sudah ada (os.MkdirAll idempotent --
// tidak error kalau folder sudah exist), dan menghindari asumsi folder
// sudah pasti dibuat sebelumnya.
func (s *LocalStorage) Save(filename string, content []byte) error {
	if err := os.MkdirAll(s.baseDir, 0o755); err != nil {
		return fmt.Errorf("failed to create upload directory: %w", err)
	}

	fullPath := filepath.Join(s.baseDir, filename)
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

// Delete menghapus file dari disk. Kalau file-nya SUDAH TIDAK ADA
// (misal terhapus manual), ini TIDAK dianggap error -- tujuan akhirnya
// ("file itu tidak ada di disk") sudah tercapai baik lewat kita hapus
// atau memang sudah tidak ada sejak awal.
func (s *LocalStorage) Delete(filename string) error {
	fullPath := filepath.Join(s.baseDir, filename)
	err := os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
