# Update delta (bsdiff) — riset dan rencana

Status: **diimplementasikan di v1.8.0** (cara CI membuat dan menandatangani
patch ada di docs/releasing.md). Dokumen ini mencatat pengukuran pada rilis
nyata repo ini, jadi keputusannya berdasar angka, bukan perkiraan.

## Masalahnya

Update hari ini mengunduh exe penuh: `WhatsUpClients-windows-amd64.exe` 42,9 MB
(Windows) dan `WhatsUpClients-linux-amd64` 43,2 MB (Linux). Exe itu adalah
aplikasi Go statis dengan font dan data emoji di dalamnya; sebagian besar isinya
sama antar rilis. Menyalin ulang 43 MB untuk satu perbaikan kecil itu mahal.

## Yang diukur

Perangkat: Windows, 8 core, Go 1.27. Sumber: empat aset rilis asli
(`gh release download`), jadi byte yang diukur persis byte yang diterima
pengguna. Semua patch diverifikasi dengan menerapkannya kembali dan
membandingkan hasilnya (`cmp`) — byte-identik dengan exe rilis tujuannya.

| Pasangan rilis | Exe baru | Full (unduhan hari ini) | gzip -9 | xz -9 | zstd -19 patch-from | bsdiff |
| --- | --- | --- | --- | --- | --- | --- |
| v1.2.2 → v1.2.3 | 42,31 MB | 42,31 MB | 22,1 MB | 18,7 MB | 62 KB | **350 B** |
| v1.4.1 → v1.5.0 | 42,94 MB | 42,94 MB | — | — | 3,14 MB | **1,34 MB** |
| v1.2.3 → v1.4.1 | 42,72 MB | 42,72 MB | — | — | 3,86 MB | **1,60 MB** |
| v1.2.2 → v1.5.0 | 42,94 MB | 42,94 MB | — | — | 4,14 MB | **1,97 MB** |

Temuan penting: ukuran patch hampir tidak tumbuh dengan jarak versi. Dari rilis
bersebelahan ke sepuluh rilis terpisah (v1.2.2 → v1.5.0, hampir seluruh riwayat
repo) patch bsdiff naik dari 1,34 MB ke 1,97 MB saja. Yang dibayar pengguna
adalah pergeseran tata letak kode Go antar build, bukan banyaknya fitur baru.

Biaya (diukur, peak "Go heap" dari sampling `runtime.MemStats`):

| Langkah | bsdiff (kr/binarydist) | bsdiff (gabstv) | zstd patch-from (klauspost) |
| --- | --- | --- | --- |
| Membuat patch | 22–33 s, 727–775 MB | 28 s, 978 MB | 2,2–3,0 s, 406–428 MB |
| Menerapkan patch | 0,6–0,9 s, 123–126 MB | 0,4 s, 108 MB | 0,05–0,1 s, 81–85 MB |
| Tambahan ukuran exe | +98 KB | +135 KB | +588 KB |

Angka terakhir diukur dengan membangun program Go kecil (baseline 1.650.688 byte,
`-trimpath -ldflags "-s -w"`) yang hanya memakai applier masing-masing, jadi
itulah biaya nyata menambahkan dekoder ke aplikasi.

Untuk perbandingan, mengompresi unduhan penuh tanpa delta: `gzip -9` 22,1 MB,
`bzip2 -9` 21,2 MB, `xz -9` 18,7 MB, `zstd -19` 20,8 MB. Jadi cara paling murah
mengurangi unduhan (tanpa delta) hanya memotong sekitar setengah.

## Kesimpulan

**Delta layak, dan bsdiff adalah pilihannya.** Untuk update biasa pengguna
mengunduh ~1,3–2 MB, bukan 42,9 MB: sekitar 20–30× lebih ringan di kasus umum,
dan ~120.000× di rilis yang perubahannya kecil. zstd patch-from juga bekerja
(62 KB vs 350 B pada rilis kecil, 3,14 MB vs 1,34 MB pada rilis fitur), dengan
pembuatan 10× lebih cepat dan memori separuh — tapi patchnya 1,4–2,4× lebih
besar dan appliernya menambah 588 KB ke exe, bukan 98–135 KB.

Alasan memilih bsdiff:

- Formatnya beku dan sederhana (BSDIFF40, Colin Percival, 2003; dulu bernama
  `bdiff`/`bpatch`, diganti karena nama itu sudah dipakai program lain).
- Ada implementasi Go murni untuk **kedua** sisi, jadi satu dependensi, tanpa
  cgo — cocok dengan build Windows yang `CGO_ENABLED=0`. `gabstv/go-bsdiff`
  memakai bzip2 Go murni (`dsnet/compress`), sehingga CI tidak perlu proses
  `bzip2` eksternal; `kr/binarydist` memakai CLI `bzip2` saat membuat patch dan
  `compress/bzip2` standar saat menerapkannya.
- Biaya pembuatan hanya dibayar sekali di CI (30 s, <1 GB di runner 7 GB);
  yang tersentuh pengguna hanya apply: 0,4–0,9 s dan ~110–126 MB sesaat.

Bukan pilihan di sini:

- **xdelta3.** Format VCDIFF-nya bagus dan memorinya berjendela, tapi encoder Go
  murni tidak ada, dan applier butuh cgo atau membawa `xdelta3.exe` — bertabrakan
  dengan build Windows tanpa cgo. Percival sendiri melaporkan patch bsdiff
  50–80% lebih kecil dari Xdelta pada uji exe. Tidak diukur di sini karena mesin
  ini tidak punya kompiler C.
- **msdelta (Windows)** — API OS yang bisa dipanggil lewat `syscall` tanpa cgo,
  tapi hanya Windows, formatnya tidak terdokumentasi, dan pembuatannya harus di
  runner Windows.
- **Courgette/Zucchini** — disassembly-aware, hasilnya spektakuler di Chrome,
  tapi khusus format exe mereka dan bukan dependensi yang wajar untuk proyek ini.

Perhatikan juga: `bdiff` di luar konteks bsdiff menunjuk ke alat lain (mis.
`bdiff`/`bpatch` Stefan Reuther dari dunia Pascal/Windows, dan `bdiff` Solaris
untuk teks). Formatnya tidak sama dengan BSDIFF40 dan tidak ada implementasi Go
untuknya, jadi jangan dipakai sebagai target kompatibilitas.

## Rencana

### Aset rilis

```
WhatsUpClients-windows-amd64.exe                        # exe penuh (sudah ada)
WhatsUpClients-windows-amd64-<sha256 sumber 16 hex>.bsdiff
WhatsUpClients-linux-amd64-<sha256 sumber 16 hex>.bsdiff
```

Nama aset memuat hash exe **sumber**, jadi patch hanya cocok dengan byte exe
yang persis. Klien tahu hash exe-nya sendiri (baca file yang sedang berjalan;
`sha256sum` atas exe 43 MB di mesin ini 0,24 s), jadi pemilihan patch tidak
butuh manifest baru: cukup cari aset yang namanya memuat hash 16 karakter itu.

Job release membuat patch dari K rilis terakhir (`gh release download` untuk
aset lama, lalu jalan). Karena ukurannya nyaris tidak tumbuh dengan jarak, K
bisa besar — misalnya 10 rilis: 20 aset kecil, ~20 MB. Batas GitHub 1000 aset
per rilis, jadi tidak jadi soal.

### Verifikasi tanpa kunci baru

Rantainya sudah kuat dengan bahan yang ada:

1. SHA256SUMS memuat hash file `.bsdiff` **dan** hash exe tujuan, keduanya
   ditandatangani ed25519 seperti sekarang. Patch tidak perlu tanda tangan
   sendiri.
2. Klien memverifikasi hash patch terhadap SHA256SUMS yang bertanda tangan
   sebelum menerapkannya (batas ukuran patch, mis. 16 MB).
3. Setelah diterapkan, klien memeriksa panjang hasil dan SHA-256 hasil terhadap
   entri exe di SHA256SUMS yang bertanda tangan itu juga.

Akibatnya patch palsu selalu gagal — patch asli untuk sumber yang salah pun
menghasilkan byte lain dan ditolak — sementara patch tetap tidak menambah materi
kunci baru. Kunci publik tidak berubah, jadi aplikasi versi lama tidak perlu
memasang rilis khusus untuk mulai memakai delta.

### Alur klien

1. Hash sekali exe yang sedang berjalan.
2. `Check`: cari aset `.bsdiff` yang cocok di rilis terbaru. Kalau tidak ada
   (versi sumber lebih tua dari K, exe diubah tangan, build dev) → jalur lama:
   unduh exe penuh. Tidak ada jalur baru yang bisa membuat update gagal total.
3. `Download`: unduh patch, verifikasi hash, terapkan dengan exe yang berjalan
   sebagai basis, verifikasi hash hasil, tulis ke `Staged`.
4. `Apply` dan restart seperti sekarang. Teks di Settings > Help berubah sendiri
   karena ukurannya diambil dari `Release.Size()` ("Download (1,3 MB) and
   restart").

### Perubahan kode

- `internal/update`: tipe patch di `Release`, pemilihan aset, jalur unduh +
  terapkan, verifikasi hasil.
- `cmd/mkpatch` (baru): pembuat patch untuk CI, memakai pustaka yang sama.
- `.github/workflows/build.yml`: langkah pembuatan patch di job release,
  hash-nya masuk `SHA256SUMS` **sebelum** ditandatangani.
- Satu dependensi baru di `go.mod` (bsdiff Go murni); exe membesar ~100–135 KB.

### Catatan memori

Apply menahan exe lama (43 MB), patch, dan hasil (43 MB) sekaligus: puncak
~110–126 MB **sesaat**, lalu dilepas. Itu di luar steady state aplikasi, tapi
karena proyek ini mengejar RAM rendah, angkanya perlu tercatat dan kalau perlu
dibatasi (tolak patch/`NewSize` yang di luar ukuran exe yang ditandatangani).

## Cara mengulang pengukuran

```sh
mkdir -p /tmp/deltalab && cd /tmp/deltalab
for tag in v1.4.1 v1.5.0; do
  gh release download "$tag" -R latiefahmad/whatsupclients \
    -p 'WhatsUpClients-windows-amd64.exe' --output "$tag.exe"
done
# harness Go kecil: klauspost/compress zstd (WithEncoderDictRaw) dan
# gabstv/go-bsdiff / kr/binarydist; ukur waktu, peak heap, lalu cmp hasilnya.
```

Yang belum diukur di sini: xdelta3, HDiffPatch, Courgette (butuh toolchain
C/C++), Linux/arm64, dan rilis macOS (belum ada di workflow).
