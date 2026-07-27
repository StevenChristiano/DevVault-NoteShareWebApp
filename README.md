# DevNote Backend — Tahap 1: Foundation & Database

## Struktur folder (Clean Architecture)

```
devnote-backend/
├── cmd/api/main.go          # entrypoint, cuma "merangkai" komponen
├── internal/
│   ├── config/              # baca .env -> struct Config
│   ├── entity/               # 8 model/struct = 8 tabel
│   ├── repository/          # (kosong, diisi Tahap 2+) akses data
│   ├── usecase/             # (kosong, diisi Tahap 2+) logika bisnis
│   └── delivery/http/       # (kosong, diisi Tahap 2+) handler & routing
├── pkg/database/            # koneksi GORM + AutoMigrate
└── .env.example
```

## Cara jalanin

1. **Ganti module path.** Semua file pakai
   `github.com/YOUR_USERNAME/devnote-backend`. Ganti `YOUR_USERNAME` di
   `go.mod` DAN di semua import (`main.go`, `pkg/database/*.go`) sesuai
   nama GitHub kamu, atau pakai nama bebas kalau belum mau push ke GitHub
   (mis. `devnote-backend` saja tanpa domain).

   Cara cepat (Linux/Mac), dari dalam folder project:
   ```bash
   grep -rl "YOUR_USERNAME" . | xargs sed -i 's/YOUR_USERNAME/nama-kamu/g'
   ```

2. **Buat database kosong lewat DBeaver.**

   PENTING dipahami: GORM `AutoMigrate` cuma membuat **tabel** di dalam
   database yang sudah ada. Dia TIDAK bisa membuat database itu sendiri —
   itu harus kamu buat manual dulu sekali lewat DBeaver:

   - Di DBeaver, pastikan kamu sudah punya koneksi ke PostgreSQL server
     kamu (biasanya `localhost`, port `5432`, dengan user `postgres` dan
     password yang kamu set waktu install PostgreSQL).
   - Klik kanan koneksi itu → **Create New Database**, kasih nama
     `devnote`, lalu Create/OK.
   - Selesai — tidak perlu bikin tabel manual, itu tugas `AutoMigrate`
     pas `main.go` dijalankan.

3. **Copy env & sesuaikan dengan kredensial PostgreSQL kamu:**
   ```bash
   cp .env.example .env
   ```
   Buka `.env`, isi `DB_USER` dan `DB_PASSWORD` sesuai yang kamu pakai di
   DBeaver waktu connect ke PostgreSQL (bukan default `postgres`/`postgres`
   kalau kamu set password lain waktu install). `DB_NAME` biarkan `devnote`
   (harus sama persis dengan nama database yang kamu buat di langkah 2).

4. **Install dependency & jalankan:**
   ```bash
   go mod tidy
   go run cmd/api/main.go
   ```

5. **Cek hasilnya:**
   - Terminal harus menampilkan log SQL `CREATE TABLE ...` untuk 8 tabel
     (users, notes, note_accesses, likes, saved_notes, follows,
     attachments, video_bookmarks — semua plural, konvensi default GORM),
     lalu `✅ auto-migrate selesai, 8 tabel siap`.
   - Buka `http://localhost:8080/health` di browser, harus muncul
     `{"status":"ok","env":"development"}`.
   - Buka DBeaver, refresh koneksinya (klik kanan → Refresh), masuk ke
     database `devnote` → Schemas → `public` → Tables. Harus muncul 8
     tabel: `users`, `notes`, `note_accesses`, `likes`, `saved_notes`,
     `follows`, `attachments`, `video_bookmarks` — lengkap dengan kolom
     dan foreign key sesuai dokumen teknis.

## Lanjut ke Tahap 2

Tahap 2 (Authentication & Access Control Middleware) akan mengisi:
- `internal/repository/user_repository.go`
- `internal/usecase/auth_usecase.go`
- `internal/delivery/http/auth_handler.go` + middleware JWT
