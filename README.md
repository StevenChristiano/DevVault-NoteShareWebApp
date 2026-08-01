# devvault Backend 
- Phase 1: Foundation & Database
- Phase 2: Authentication & Access Control Middleware
- Phase 3: Core Features - Notes, Youtube & File Upload

## Struktur folder (Clean Architecture) (Updated Phase 2)

```
devvault-backend/
├── cmd/api/main.go          # entrypoint, cuma "merangkai" komponen   # (Covered in Phase 1-4(as long as adding new module))
├── internal/
│   ├── config/              # baca .env -> struct Config              # (Covered in Phase 1)
│   ├── entity/              # 8 model/struct = 8 tabel                # (Covered in Phase 1)
│   ├── repository/          # akses data (query GORM)                 # (Covered in Phase 2 +)
│   ├── usecase/             # logika bisnis                           # (Covered in Phase 2 +)
│   └── delivery/
│       └── http/                                                      # (Covered in Phase 2)
│           ├── middleware/      # AuthMiddleware, AccessMiddleware, OptionalAuthMiddleware
│           ├── auth_handler.go
│           ├── note_handler.go
│           ├── note_access_handler.go
│           └── routes.go
├── pkg/
│   ├── database/                 # koneksi GORM + AutoMigrate
│   └── token/                    # generate & verifikasi JWT          # (Covered in Phase 2)
│   └── slug/                     # slugify teks untuk URL note public
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

## Endpoint

### Auth

| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/v1/auth/register` | Body: `{"username","email","password"}`. Password min 8 karakter, wajib huruf besar+kecil+angka. |
| POST | `/api/v1/auth/login` | Body: `{"email","password"}`. Sukses → cookie `token` (HTTP-Only). |

### Notes

| Method | Endpoint | Akses | Keterangan |
|---|---|---|---|
| POST | `/api/v1/notes` | Login | Buat note baru. Body: `{"header","paragraph","visibility"}`. |
| GET | `/api/v1/notes` | Login | Daftar note milik sendiri. |
| GET | `/api/v1/notes/:slug` | Dynamic | Lihat note. Tidak wajib login kalau `public`. `view_count` bertambah tiap dibuka. |
| PUT | `/api/v1/notes/:id` | Owner/Editor | Update note (partial). Owner boleh ubah `visibility`; Editor cuma `header`/`paragraph`. |
| DELETE | `/api/v1/notes/:id` | Owner only | Hard delete, CASCADE ke data terkait. |

### Note Access (Owner only)

| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/v1/notes/:id/access` | Body: `{"email","role"}` (`role`: `viewer`/`editor`). Upsert — buat baru atau ubah role kalau sudah ada. |
| DELETE | `/api/v1/notes/:id/access` | Body: `{"email"}`. Cabut akses user tersebut dari note ini. |
| GET | `/api/v1/notes/:id/access` | Daftar semua user yang punya akses ke note ini beserta role-nya. |

### Video Bookmark (YouTube)

| Method | Endpoint | Akses | Keterangan |
|---|---|---|---|
| POST | `/api/v1/notes/:id/bookmarks` | Owner/Editor | Body: `{"youtube_url","timestamp_sec","note_text"}`. `youtube_id` diekstrak otomatis dari URL. |
| GET | `/api/v1/notes/:id/bookmarks` | Dynamic (ikut visibility note) | Daftar bookmark milik note ini, urut berdasarkan `timestamp_sec`. |
| DELETE | `/api/v1/notes/:id/bookmarks/:bookmarkId` | Owner/Editor | Hapus satu bookmark. |

## Model akses: Visibility dan Role Editor bersifat INDEPENDEN

Ini keputusan desain penting — **visibility** (siapa boleh lihat) dan **note_access** (siapa boleh edit) adalah dua hal terpisah yang **tidak saling mengubah** satu sama lain:

| Visibility | Siapa boleh **lihat** | Siapa boleh **edit** |
|---|---|---|
| `private` | Owner saja | Owner saja (note_access diabaikan total) |
| `public` | Semua orang (termasuk guest) | Owner **+** siapa pun yang di-set `editor` lewat `note_access` |
| `shared` | Owner **+** siapa pun yang **terdaftar** di `note_access` (whitelist) | Owner **+** yang role-nya `editor` |

Mengubah visibility **tidak pernah** menghapus baris `note_access` yang sudah ada — riwayat siapa viewer/editor tetap tersimpan walau visibility berubah-ubah, dan akan otomatis relevan lagi begitu visibility balik ke `public`/`shared`.



## Status roadmap

- ✅ Tahap 1: Foundation & Database
- ✅ Tahap 2: Authentication & Access Control Middleware
- 🔄 Tahap 3: Core Features — Notes, YouTube & File Upload
  - ✅ CRUD Note (Create, Read, Update, Delete + slug generation)
  - ✅ Note Access (grant/revoke/list viewer & editor lewat email)
  - ✅ YouTube Parser Helper
  - ⏭️ File Upload Helper
- ⏭️ Tahap 4: Social Features — Like, Save, Follow & FYP
- ⏭️ Tahap 5: Frontend Next.js
