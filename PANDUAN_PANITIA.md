# Panduan Panitia — Kejuaraan Beregu Internal PTM Laguna

Cibubur, Jumat–Sabtu 6–7 November 2026 · 6 tim (A–F) · setengah kompetisi · 15 pertandingan

Tidak perlu install apa pun dan tidak perlu login. Cukup buka link di bawah dari browser
(Chrome di HP/tablet, browser bawaan smart TV, atau laptop) yang tersambung ke internet.

## Link

| Untuk | Link |
|---|---|
| Menu utama (semua link + QR) | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna |
| **Layar TV** | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna#tv |
| **Input skor meja 1** | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna#meja1 |
| **Input skor meja 2** | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna#meja2 |
| **Panitia / admin** | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna#admin |
| Uji kecepatan internet | https://brot1994.github.io/SkorTenisMeja/?k=ptmlaguna#uji |

- Kode ruang: **ptmlaguna**
- PIN panitia: **ptmlaguna** (diminta sekali saat pertama kali mengubah skor atau pengaturan di perangkat itu)

Perangkat yang sudah pernah membuka link akan ingat ruangnya, jadi berikutnya cukup buka halaman yang sama dari riwayat browser.

---

## 1. Layar TV

1. Buka link **Layar TV** di browser TV, lalu jadikan layar penuh.
2. TV hanya menampilkan, tidak perlu disentuh lagi. Isinya:
   - Tabel silang 6 tim: skor partai tiap pertemuan, titik hijau = partai yang dimenangkan, kotak merah = sedang main.
   - Kolom M-K (menang-kalah), selisih partai, poin (menang = 1, kalah = 0), dan ranking.
   - Kartu Meja 1 dan Meja 2: skor langsung, partai keberapa, nama pemain.
   - Daftar **Berikutnya**: pertandingan yang siap dipanggil, atau "Istirahat s/d jam …" kalau timnya baru selesai main.
3. Kalau TV terputus dari internet, halaman akan tersambung lagi sendiri. Kalau macet, cukup refresh.

## 2. Wasit meja (input skor)

Satu perangkat per meja: tablet, HP, atau laptop. Buka link **Input skor meja 1** atau **meja 2**.

**Memulai pertandingan**
1. Saat meja kosong, muncul daftar pertandingan. Urutan teratas adalah saran sistem (sudah memperhitungkan istirahat 10 menit).
2. Tekan **Mulai** di pertandingan yang akan dimainkan.
3. Muncul form **Susunan pemain** untuk ketiga partai sekaligus. Pilih nama dari daftar pemain tiap tim:
   Tunggal 1 (1 nama per tim), Ganda (2 nama per tim), Tunggal 2 (1 nama per tim).
   Kalau satu pemain dipilih di dua partai tunggal, atau nama yang sama dipilih dua kali di ganda, muncul peringatan kuning.
4. Tekan **Mulai pertandingan**. Kalau ada yang belum diisi, tombolnya menjadi **Mulai (sisanya diisi nanti)**. Kalau diminta PIN, isi PIN panitia.

**Mengganti pemain di tengah pertandingan**
- Di bawah tab partai (Tunggal 1 · Ganda · Tunggal 2) ada pilihan pemain untuk partai yang sedang berjalan.
  - Tunggal: pilih **1 nama** per tim.
  - Ganda: pilih **2 nama** per tim.
- Setiap nama ditampilkan dengan hitungan main, misalnya `Eny (T1 G0)` = sudah 1× tunggal, belum pernah ganda.
- Tanda **★** = pemain itu belum pernah main di jenis partai yang sedang dipilih.
- **Aturan: semua pemain minimal sekali main tunggal dan sekali main ganda.** Prioritaskan nama yang bertanda ★.
- Pemain bisa dipilih atau diganti kapan saja selama partai berjalan, skor tidak hilang.

**Mencatat angka**
- Tekan **+1** di sisi tim yang mendapat angka. Salah tekan: **−1** atau **Undo**.
- Di laptop bisa pakai keyboard: `←` / `A` = sisi kiri, `→` / `L` = sisi kanan, `Z` / `Backspace` = undo.
- **Tukar sisi** kalau pemain berganti sisi meja (hanya tampilan di perangkat ini).
- Set selesai otomatis di 11 (harus unggul 2, deuce ditandai). Best of 3 set.
- Partai pindah otomatis: Tunggal 1 → Ganda → Tunggal 2. Pertandingan selesai begitu satu tim menang 2 partai.

**Setelah selesai**
- Tekan **Kosongkan meja, lanjut partai berikutnya** (atau `Enter` di laptop), lalu pilih pertandingan berikutnya.

**Tidak jadi main / pemain belum datang**
- Ganti pemainnya saja di pilihan pemain, atau
- Tekan **Batalkan pertandingan** (tombol merah). Skor dihapus, meja kosong, dan pertandingan kembali ke antrean untuk dimainkan nanti.

## 3. Panitia / admin

Buka link **Panitia / admin**.

- **Rekap main pemain**: T (tunggal) dan G (ganda) setiap pemain. Merah = belum pernah. Kalau tim sudah habis pertandingannya tapi ada pemain yang belum memenuhi aturan, muncul peringatan.
- **Pertandingan**: daftar 15 pertandingan dengan status dan hasilnya. Tekan **Ubah** untuk:
  - memperbaiki skor (tulis per set, misalnya `11-8, 9-11, 11-7`, angka kiri = tim pertama),
  - memperbaiki nama pemain per partai,
  - **Kosongkan (belum main)** untuk menghapus hasil dan mengembalikannya ke antrean.
- **Aturan dan meja**: jumlah meja, istirahat minimal, best of per babak. Tekan **Simpan perubahan** setelah mengubah.
- **Peserta**: nama tim, warna, dan daftar pemain bisa diubah di bagian Beregu → Peserta.
- **Tandai turnamen selesai** setelah pertandingan terakhir. Data tetap tersimpan di **Arsip**.
- **Unduh backup (.json)**: simpan cadangan sesekali, misalnya di akhir hari pertama.

Beberapa admin boleh membuka halaman panitia bersamaan. Perubahan langsung terlihat di semua perangkat.

### Mengganti pemain (ada yang tidak bisa datang)

**A. Hanya untuk satu pertandingan**, pemain pengganti sudah ada di daftar tim:
- Saat **Mulai**, pilih nama lain di form **Susunan pemain**.
- Kalau partai sudah berjalan, ganti di baris pilihan pemain di bawah tab partai. Skor tidak hilang.

**B. Mengganti atau menambah orang di daftar tim**, misalnya ada pemain baru yang belum terdaftar:
1. Buka **Panitia** → kartu **Beregu** → klik **Peserta (6): ubah nama, divisi, warna, pemain**.
2. Di kolom pemain tim yang bersangkutan, ubah atau tambahkan nama, dipisahkan koma.
   Contoh: `Agung, Eny, Arie, Tantri, Jose, Rudi`
3. Gulir ke bawah, tekan **Simpan perubahan**.
4. Nama baru langsung muncul di pilihan pemain di semua tablet/HP wasit.

Catatan: pertandingan yang sudah selesai tetap mencatat nama yang benar-benar main saat itu. Kalau nama lama
dihapus dari daftar setelah ia sempat main, namanya tetap muncul di **Rekap main pemain** beserta hitungan T/G-nya.

## 4. Kalau ada masalah

| Masalah | Coba |
|---|---|
| Tulisan "Menyambung ulang…" | Cek internet perangkat itu. Angka yang ditekan tetap disimpan dan dikirim otomatis setelah tersambung. |
| TV tidak berubah | Refresh browser TV. |
| Diminta PIN terus | Pastikan PIN diketik huruf kecil semua: `ptmlaguna`. |
| Masuk ke halaman "Ruang turnamen" | Buka lagi link lengkap yang ada `?k=ptmlaguna`, atau ketik `ptmlaguna` di kolom Kode ruang lalu **Masuk ke ruang ini**. |
| Salah tekan Mulai | **Batalkan pertandingan** di halaman input skor. |
| Skor sudah selesai tapi salah | Panitia → Pertandingan → **Ubah**. |
| Pemain berhalangan, diganti orang lain | Lihat bagian **Mengganti pemain** di atas. |

## Catatan keamanan

Link dan PIN di panduan ini terbuka untuk siapa saja yang membaca halaman ini. Kalau panduan ini akan
dibagikan luas (misalnya ke grup WA peserta), sebaiknya ganti PIN di **Panitia → Backup dan keamanan**
dan bagikan PIN baru hanya ke panitia.
