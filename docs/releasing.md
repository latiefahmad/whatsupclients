# Build otomatis dan rilis

Workflow `.github/workflows/build.yml` berjalan pada setiap push branch, pull
request, push tag `v*`, dan melalui **Actions > Build and release > Run workflow**.
Pilihan manual tersedia setelah workflow masuk ke default branch.

Versi Go mengikuti `go.mod`. Kedua target menerapkan patch go-text
(`patches/apply.sh`), lalu menjalankan pemeriksaan format, `go mod verify`,
`go vet ./...`, `go test ./...`, dan `go build ./...` sebelum membangun aplikasi:

| File | Isi |
| --- | --- |
| `WhatsUpClients-Setup.exe` | Installer Windows (Inno Setup, `installer/whatsup.iss`) |
| `WhatsUpClients-windows-amd64.exe` | Aplikasi Windows portable, tanpa jendela console; juga file yang diunduh oleh tombol update |
| `WhatsUpClients-linux-amd64` | Executable Linux |
| `SHA256SUMS`, `SHA256SUMS.sig` | Checksum semua file di atas, dan tanda tangan ed25519-nya (hanya di release) |

Unduh hasil dari bagian **Artifacts** pada run yang berhasil. Artefak disimpan
selama 14 hari; aset yang sudah dilampirkan ke GitHub Release tetap tersimpan.

Linux dibangun di Ubuntu 22.04 dengan CGO serta dukungan X11/Wayland. Paket ini
bukan binary statis atau AppImage; komputer tujuan memerlukan lingkungan desktop
dan library runtime Gio (X11/Wayland, xkbcommon, EGL/GLES). Lihat
[dependensi Linux Gio](https://gioui.org/doc/install/linux). macOS, ARM64, dan
code signing Windows belum disediakan oleh workflow ini.

## Installer Windows

Installer memasang aplikasi untuk user yang sedang login saja, ke
`%LocalAppData%\Programs\WhatsUpClients`, tanpa meminta hak admin. Installer membuat
shortcut di Start Menu (opsional di Desktop) dan mendaftarkan uninstaller di
**Settings > Apps**. Uninstaller menutup aplikasi, menghapus entri registry yang
dibuat aplikasi sendiri (notifikasi dan start at login), lalu menanyakan apakah
chat dan sesi WhatsApp (`%AppData%\WhatsUpClients`) ikut dihapus. Defaultnya tidak.

Untuk membangunnya secara lokal, pasang [Inno Setup 6](https://jrsoftware.org/isinfo.php),
build exe ke `bin/dist/WhatsUpClients-windows-amd64.exe`, lalu jalankan
`iscc /DAppVersion=1.0.0 installer\whatsup.iss`.

## Update dari dalam aplikasi

**Settings > Help** menampilkan versi aplikasi dan tombol **Check for updates**.
Aplikasi juga memeriksa sendiri dengan diam-diam: sekali sesaat setelah
dijalankan, lalu sehari sekali. Pemeriksaan hanya mengambil metadata rilis
GitHub (tanpa data pengguna); unduhan hanya terjadi kalau pengguna memilih
**Update now**. Kalau ada versi yang lebih baru, aplikasi menawarkan dialog
**Update now** atau **Later**. **Later** melewatkan versi itu sampai pengguna
memeriksa sendiri dari Settings > Help, sedangkan tombol Help itu sendiri
mengunduh exe untuk sistem ini, mencocokkannya dengan `SHA256SUMS`, memeriksa
tanda tangan `SHA256SUMS.sig` dengan kunci publik di `internal/update`,
menukar exe yang sedang berjalan, lalu me-restart aplikasi.

- Hanya build dari tag (`-X main.version=v1.2.3`) yang bisa update. Build lain,
  dan prerelease seperti `v1.2.3-rc1`, adalah development build.
- Yang dicek adalah *latest release* GitHub, jadi draft dan prerelease tidak
  pernah ditawarkan.
- Update butuh folder exe yang bisa ditulis. Installer memasang ke folder seperti
  itu; untuk versi portable, letakkan exe di folder milik user.

### Kunci tanda tangan

Job release menandatangani `SHA256SUMS` dengan `cmd/signrelease` memakai secret
`RELEASE_SIGNING_KEY` (seed ed25519 dalam base64). Tanpa secret itu job release
gagal. Kunci publiknya ada di `publicKey` (`internal/update/update.go`).

Isi secret dengan **isi file hasil `-keygen` (seed privat)**, bukan kunci
publik yang dicetaknya. Keduanya sama-sama 32 byte base64, jadi kunci publik
yang kepaste pun lolos dan menandatangani dengan kunci yang salah — tanpa ada
yang gagal sampai pengguna tidak bisa update. CI mencegahnya: step Sign
checksums membandingkan kunci hasil sign dengan `publicKey` di kode dan gagal
kalau beda. Ritual tiap rilis: pastikan baris log `signed ... with public key`
sama dengan `publicKey` di kode.

Simpan cadangan kunci privat di tempat aman. Kalau kunci hilang atau bocor, buat
kunci baru (`go run ./cmd/signrelease -keygen <file>`, yang mencetak kunci publik
barunya), ganti `publicKey` dan secret-nya. Versi lama aplikasi tidak akan mau
memasang rilis yang ditandatangani kunci baru, jadi penggunanya harus mengunduh
satu rilis itu secara manual.

## Membuat rilis

1. Tulis section versi baru di `CHANGELOG.md` (paling atas, `## [vX.Y.Z] - YYYY-MM-DD`).
   Workflow menerbitkannya verbatim sebagai release notes, dan gagal kalau
   section-nya tidak ada — jadi changelog selalu ditulis sebelum tag.
2. Regenerasi snapshot landing page: `node build-releases.js`
   (menulis `releases.js`, sumber data section *Rilis* di halaman;
   tanpa ini LP tetap menampilkan versi lama).
3. Commit dan push source beserta workflow ke repository GitHub.
4. Pada commit yang akan dirilis, buat dan push tag versi baru, misalnya:

   ```sh
   git tag -a v1.0.0 -m "WhatsUpClients v1.0.0"
   git push origin v1.0.0
   ```

5. Setelah kedua build berhasil, workflow langsung membuat **release yang
   dipublikasikan** dengan semua file di atas dan isi `CHANGELOG.md` sebagai
   release notes. Aplikasi menawarkannya sebagai update begitu rilis terbit.
   Push tag berarti langsung mengirim update ke pengguna, jadi periksa dulu
   sebelum push tag.

Hanya job release yang memiliki izin `contents: write`. Kebijakan
repository/organisasi harus mengizinkan GitHub Actions membuat release.

Menjalankan ulang run tag tidak menimpa aset yang sudah terbit (workflow
menolak mengganti aset release yang published); gunakan tag versi baru.
Run manual dan push branch hanya menghasilkan artefak, tanpa membuat release.

Untuk memeriksa unduhan di Linux, simpan file dan `SHA256SUMS` dalam satu
direktori, lalu jalankan `sha256sum --check --ignore-missing SHA256SUMS`. Di
PowerShell, gunakan `Get-FileHash .\WhatsUpClients-Setup.exe -Algorithm SHA256`
dan cocokkan hasilnya dengan entri di `SHA256SUMS`.
