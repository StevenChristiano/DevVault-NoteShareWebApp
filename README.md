# devvault Backend 
- Phase 1: Foundation & Database
- Phase 2: Authentication & Access Control Middleware 

## Struktur folder (Clean Architecture) (Updated Phase 2)

```
devvault-backend/
├── cmd/api/main.go          # entrypoint, cuma "merangkai" komponen   # (Covered in Phase 1)
├── internal/
│   ├── config/              # baca .env -> struct Config              # (Covered in Phase 1)
│   ├── entity/               # 8 model/struct = 8 tabel               # (Covered in Phase 1)
│   ├── repository/          # (kosong, diisi Tahap 2+) akses data     # (Covered in Phase 2 +)
│   ├── usecase/             # (kosong, diisi Tahap 2+) logika bisnis  # (Covered in Phase 2 +)
│   └── delivery/
│       └── http/                                                      # (Covered in Phase 2)
│           ├── middleware/      # AuthMiddleware, AccessMiddleware
│           ├── auth_handler.go
│           └── routes.go
├── pkg/
│   ├── database/                 # koneksi GORM + AutoMigrate
│   └── token/                    # generate & verifikasi JWT          # (Covered in Phase 2)
└── .env.example                  # env template
```

## Cara jalanin

1. **Ganti module path.** Semua file pakai
   `github.com/YOUR_USERNAME/devvault-backend`. Ganti `YOUR_USERNAME` di
   `go.mod` DAN di semua import (`main.go`, `pkg/database/*.go`) sesuai
   nama GitHub kamu, atau pakai nama bebas kalau belum mau push ke GitHub
   (mis. `devvault-backend` saja tanpa domain).

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
     `devvault`, lalu Create/OK.
   - Selesai — tidak perlu bikin tabel manual, itu tugas `AutoMigrate`
     pas `main.go` dijalankan.

3. **Copy env & sesuaikan dengan kredensial PostgreSQL kamu:**
   ```bash
   cp .env.example .env
   ```
   Buka `.env`, isi `DB_USER` dan `DB_PASSWORD` sesuai yang kamu pakai di
   DBeaver waktu connect ke PostgreSQL (bukan default `postgres`/`postgres`
   kalau kamu set password lain waktu install). `DB_NAME` biarkan `devvault`
   (harus sama persis dengan nama database yang kamu buat di langkah 2).
   ganti `JWT_SECRET` dengan string acak sendiri (jangan pakai
   nilai contoh di `.env.example`).

5. **Install dependency & jalankan:**
   ```bash
   go mod tidy
   go run cmd/api/main.go
   ```

6. **Cek hasilnya:**
   - Terminal harus menampilkan log SQL `CREATE TABLE ...` untuk 8 tabel
     (users, notes, note_accesses, likes, saved_notes, follows,
     attachments, video_bookmarks — semua plural, konvensi default GORM),
     lalu `✅ auto-migrate selesai, 8 tabel siap`.
   - Buka `http://localhost:8080/health` di browser, harus muncul
     `{"status":"ok","env":"development"}`.
   - Buka DBeaver, refresh koneksinya (klik kanan → Refresh), masuk ke
     database `devvault` → Schemas → `public` → Tables. Harus muncul 8
     tabel: `users`, `notes`, `note_accesses`, `likes`, `saved_notes`,
     `follows`, `attachments`, `video_bookmarks` — lengkap dengan kolom
     dan foreign key sesuai dokumen teknis.
   - Cek dengan Postman/Thunder Client:
     - `POST /api/v1/auth/register` body JSON:
       `{"username":"budi","email":"budi@mail.com","password":"password123"}`
       → harus dapat `201 Created`.
     - `POST /api/v1/auth/login` body JSON:
       `{"email":"budi@mail.com","password":"password123"}`
       → harus dapat `200 OK` + cookie `token` ter-set (cek tab
       Cookies di Postman).


## Status roadmap

- ✅ Tahap 1: Foundation & Database
- ✅ Tahap 2: Authentication & Access Control Middleware
  - Register & login (bcrypt untuk hash password)
  - `AuthMiddleware` (verifikasi JWT dari cookie)
  - `AccessMiddleware` (cek visibility + role note_access — siap dipakai,
    baru benar-benar dipasang ke route mulai Tahap 3)
- ⏭️ Tahap 3: Core Features — Notes, YouTube & File Upload
