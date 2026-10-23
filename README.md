# Skor Tenis Meja

Papan skor turnamen tenis meja: layar TV, input skor per meja (tablet/HP/laptop), dan halaman panitia
(turnamen baru, undian berbasis divisi, arsip, koreksi skor).

## Versi online (GitHub Pages + Firebase)

Alamat: **https://brot1994.github.io/SkorTenisMeja/**

1. Buka alamat di atas, tekan **Buat ruang baru** (atau isi kode ruang sendiri).
2. Halaman Menu menampilkan link + QR untuk:
   - Layar TV: `...?k=KODE#tv`
   - Input skor: `...?k=KODE#meja1`, `#meja2`, dst.
   - Panitia: `...?k=KODE#admin`
   - Uji delay: `...?k=KODE#uji`
3. Semua perangkat cukup terhubung ke internet (WiFi apa pun atau data seluler).

Data disimpan di Firebase Realtime Database proyek `agi-flipclock`, di bawah jalur `/turnamen/<kode ruang>`,
terpisah dari data Flip Clock (`/clocks`).

### Rules Firebase

Kalau muncul pesan *Permission denied*, buka Firebase Console → Realtime Database → **Rules**
dan tambahkan blok `turnamen` di dalam `"rules"` (biarkan blok lain seperti `clocks` tetap ada):

```json
"turnamen": { ".read": true, ".write": true }
```

### Keamanan

Siapa pun yang tahu alamat database dan kode ruang bisa membaca data. Perubahan skor dan pengaturan
dilindungi **PIN panitia** (atur di Panitia → Backup dan keamanan). PIN ini pengaman ringan untuk acara warga,
bukan pengamanan tingkat tinggi.

## Versi lokal (tanpa internet)

Folder `server-lokal/` berisi server Go (Windows `.exe` / Android) yang menyajikan halaman yang sama
lewat WiFi lokal. Folder `android/` berisi sumber APK yang menjalankan server lokal itu di tablet/TV Android.
