# Changelog

All notable changes to WhatsUp Clients are documented in this file.

The **Release** workflow (`.github/workflows/build.yml`) publishes the section
for the tagged version to GitHub **verbatim** as the release notes. Conventions:

- One section per release, headed `## [vX.Y.Z] - YYYY-MM-DD` (newest first).
- Everything under the heading until the next `## [` heading is published as-is.
- Write for end users: what is new, what is fixed, and which file to download.

## [v1.3.0] - 2026-10-09

### 😀 Reaksi

- **Ringkasan reaksi** — pil di bawah pesan menampilkan sampai tiga emoji
  terbanyak plus jumlahnya. Klik untuk melihat siapa memberi apa (satu tab
  per emoji); klik barisanmu sendiri untuk menarik reaksimu.

### 💬 Chat

- **System messages di chat** — anggota keluar/masuk, ganti nama, ikon dan
  deskripsi grup, timer disappearing, pesan PIN, link undangan, ganti kode
  keamanan dan panggilan tak terjawab tampil sebagai chip abu di tengah,
  sekaligus jadi pesan terakhir di daftar chat.
- **Tombol kembali ke terbaru** — tombol ⌄ di kanan bawah saat kamu scroll
  ke atas, lengkap dengan jumlah pesan belum dibaca di bawahnya; tanggal
  menempel di atas selama scroll.
- **"Dihapus admin *Nama*"** — pesan yang dihapus admin grup kini menyebut
  siapa yang menghapus.

### 📋 Daftar chat

- **Chip filter** — daftar custom tampil sebagai chip (Favourites, Groups,
  buatanmu) dengan **+** untuk membuat baru.
- **Search yang lebih luas** — menemukan chat arsip, kontak yang belum ada
  chat-nya, dan pesan di semua chat; hanya cocok terbaru yang ditampilkan.
- **Badge timer** — chat dengan disappearing messages punya badge jam di
  fotonya, dan urutan indikator baris mengikuti WhatsApp.

### 🎙️ Media

- **Kecepatan voice message 1×/1.5×/2×** — pilihanmu diingat untuk berikutnya.
- **Volume slider** — video, status dan voice message punya pengatur volume
  yang diingat setelah restart.
- **GIF autoplay** — GIF di chat berputar sendiri selagi terlihat (mute,
  loop, maks 3 bersamaan), seperti stiker.
- **Tombol stiker** — setelah memakai tab stiker, tombol emoji komposer
  menjadi tombol stiker.

### ⚙️ Pengaturan & sinkron

- **Settings > Performance** — matikan animasi dan smooth scrolling, atur
  kapan GIF/stiker animasi berputar, plus halaman Advanced berisi angka
  memori live, tombol Free memory, dan kenop cache, GC dan CPU.
- **System default theme** — mengikuti terang/gelap sistem (sekarang default).
- **Sync dari HP** — PIN, mute, arsip dan bacaan yang diatur di HP kini
  diterapkan di sini; history lama tak lagi menimpanya.
- Buka grup besar tak lagi freeze (media info panel dihitung di background),
  logout minta konfirmasi, `//kick` mengirim `/kick` sebagai teks, tombol
  bekerja tanpa menunggu mouse gerak, dan duplikat tombol Cancel dihapus.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.2.3...v1.3.0

## [v1.2.3] - 2026-10-07

### 🌐 Landing page

- **Project website added** — a landing page in `site/` presents the app with
  a feature tour, screenshots, dark and light themes, and download buttons for
  Windows (installer and portable) and Linux that always serve the newest
  release. Its *Rilis* section shows the notes of the latest version, and a
  button scrolls straight to the downloads.
- **README brought up to date** — the 70%–200% font size, the automatic
  update offer (*Update now* or *Later*, checked every 12 hours), and the
  Donate button are now listed, and the download section describes how
  updates arrive.

The app itself is unchanged; its files are identical to v1.2.2.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.2.2...v1.2.3

## [v1.2.2] - 2026-10-07

### 🔄 Updates

- **Faster update discovery** — the silent background check now runs every 12
  hours (was 24), so a new release is offered the same day even when the app
  stays running. The offer itself is unchanged: once per version, Update now
  or Later, your choice.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.2.1...v1.2.2

## [v1.2.1] - 2026-10-07

### 🔄 Updates

- **Fixed automatic updates** — v1.1.0 and v1.2.0 could not update themselves
  because the releases were signed with a key the app does not trust. This
  release is signed correctly, so Check for updates (and the automatic offer)
  work again. No app changes besides the fix; CI now refuses to sign with a
  mismatched key.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.2.0...v1.2.1

## [v1.2.0] - 2026-10-07

### ❤️ Donate

- **Donate button in the title bar** — a heart with "Donate" next to the
  window buttons opens the donation page in your browser.

### 📝 Snippets

- **`/snippet save` without a reply** — typing `/snippet save some text`
  keeps the text itself as a new snippet (before, saving required replying
  to a message). With a reply it still saves that message under the given
  name, as before.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.1.0...v1.2.0

## [v1.1.0] - 2026-10-07

### 🔄 Update offer

- **No need to open Settings first** — the app now checks for a newer release
  shortly after start and then daily, and offers it once per version.
- **Update now** downloads the release, verifies its signature, swaps it in
  place and restarts. **Later** skips that version until you check by hand
  from Settings > Help.

### 🔍 Zoom

- **70% zoom level** — the smallest font size step is now 70% (was 80%),
  available from Ctrl+- , the zoom bubble and Settings > General > Font size.

### 📝 Releases

- **Changelog-driven release notes** — every release now ships the matching
  `CHANGELOG.md` section as its notes, so this page always tells you what is
  new. A release without a changelog entry fails instead of shipping empty.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.0.0...v1.1.0

## [v1.0.0] - 2026-10-07

First release published from this repository.

### ✨ Highlights

- **New name and home** — WazzapClients is now **WhatsUp Clients**, published
  from `github.com/latiefahmad/whatsupclients`. Upgrades keep everything:
  the old data directory, database and login are moved to the new names
  automatically on first start.
- **Automatic releases** — pushing a `v*` tag builds, signs and publishes the
  release at once, so Check for updates offers it without a manual step.
- **Update offer** — the app checks for a newer release shortly after start
  and then daily, and offers it once per version: **Update now** downloads
  and restarts, **Later** skips that version until you check by hand from
  Settings > Help.
- **70% zoom level** — the smallest font size step is now 70% (was 80%),
  available from Ctrl+- , the zoom bubble and Settings > General > Font size.

### 📦 Downloads

| Platform | File |
| --- | --- |
| Windows 10/11 (x64) | `WhatsUpClients-Setup.exe` (installer) or `WhatsUpClients-windows-amd64.exe` (portable) |
| Linux (x64) | `WhatsUpClients-linux-amd64` |

### 🔄 Updating

Install this release once from the files above. Later releases arrive through
Settings > Help > Check for updates (or the automatic offer): the app
downloads the executable for your system, checks it against the signed
`SHA256SUMS`, swaps it in place and restarts itself.
