# Changelog

All notable changes to WhatsUp Clients are documented in this file.

The **Release** workflow (`.github/workflows/build.yml`) publishes the section
for the tagged version to GitHub **verbatim** as the release notes. Conventions:

- One section per release, headed `## [vX.Y.Z] - YYYY-MM-DD` (newest first).
- Everything under the heading until the next `## [` heading is published as-is.
- Write for end users: what is new, what is fixed, and which file to download.

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
