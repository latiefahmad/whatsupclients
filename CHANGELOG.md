# Changelog

All notable changes to WhatsUp Clients are documented in this file.

The **Release** workflow (`.github/workflows/build.yml`) publishes the section
for the tagged version to GitHub **verbatim** as the release notes. Conventions:

- One section per release, headed `## [vX.Y.Z] - YYYY-MM-DD` (newest first).
- Everything under the heading until the next `## [` heading is published as-is.
- Write for end users: what is new, what is fixed, and which file to download.

## [v1.3.0] - 2026-10-09

### 😀 Reactions

- **Reaction summaries** — the pill under a message shows up to three
  emojis, the most given first, plus their counts. Click it to see who
  reacted, with a tab per emoji; click your own row to take your
  reaction back.

### 💬 Chats

- **System messages in chats** — members joining or leaving, name, icon
  and description changes, disappearing timers, pins, invite links,
  security code changes and missed calls show as grey chips down the
  middle, and as the chat list's last message.
- **Scroll-to-bottom button** — a ⌄ button at the bottom right while
  you're scrolled up, with the unread count below it; the day stays
  pinned at the top while you scroll.
- **"Deleted by admin *Name*"** — a message a group admin deleted now
  says who did it.

### 📋 Chat list

- **Filter chips** — custom lists show as chips (Favourites, Groups and
  yours) with a **+** for a new one.
- **Wider search** — finds archived chats, contacts you have no chat
  with yet, and messages in every chat; only the newest matches show.
- **Timer badge** — chats with disappearing messages get a clock badge
  on their picture, and row indicators are ordered like WhatsApp.

### 🎙️ Media

- **Voice message speed 1×/1.5×/2×** — the speed you pick is kept for
  the next ones.
- **Volume slider** — videos, statuses and voice messages share one
  volume, remembered after a restart.
- **GIF autoplay** — GIFs play while they're on screen, muted and
  looping, at most 3 at once, like stickers.
- **Sticker button** — after you use the sticker tab, the composer's
  emoji button becomes the sticker button and opens there again.

### ⚙️ Settings & sync

- **Settings > Performance** — switch off animations and smooth wheel
  scrolling, choose when GIFs and animated stickers play, plus an
  Advanced page with live memory numbers, a Free memory button, and
  cache, GC and CPU knobs.
- **System default theme** — follows your system's light or dark mode,
  now the default.
- **Synced from your phone** — pins, mutes, archives and reads made on
  the phone now apply here, and old history no longer undoes them.
- Big groups open without freezing (their media is counted in the
  background), logging out asks first, `//kick` sends `/kick` as text,
  buttons redraw without waiting for the mouse to move, and the doubled
  Cancel button is gone.

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
