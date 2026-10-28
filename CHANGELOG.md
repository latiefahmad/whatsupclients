# Changelog

All notable changes to WhatsUp Clients are documented in this file.

The **Release** workflow (`.github/workflows/build.yml`) publishes the section
for the tagged version to GitHub **verbatim** as the release notes. Conventions:

- One section per release, headed `## [vX.Y.Z] - YYYY-MM-DD` (newest first).
- Everything under the heading until the next `## [` heading is published as-is.
- Write for end users: what is new, what is fixed, and which file to download.

## [v1.9.0] - 2026-10-28

### 🔒 Status

- **Who sees your status** — a Status privacy panel (the Status page's ⋮ menu)
  picks the audience for your updates: My contacts, My contacts except…, Only
  share with…, Close friends, or an audience of your own. The setting is synced
  with your phone, so both change together.
- **Your own audiences** — make a list with a name and an emoji ("Gym 💪") and
  pick who is in it; the people in it see your name and emoji on your updates.
- **Close friends and audience badge** — a status shared with close friends or
  an audience of yours shows a chip with its emoji and name on the picture.
- **Hide a contact's statuses** — Hide or Unhide from a contact's row; hidden
  ones move under a collapsible Hidden section, ready to unhide.
- **Only the chosen audience** — a status that would reach someone outside the
  audience you picked is refused with a notice instead of posted.

### 🎙️ Voice messages

- **Record a voice message** — the composer's mic records with a live waveform,
  a pulsing dot and a timer; pause and resume, play the take back, then send it
  (as a reply too). Encoded to OGG Opus with ffmpeg; you're told when a mic or
  ffmpeg is missing.

### 🧩 New message types and tools

- **Camera** — Attach > Camera opens your system's camera app (Windows Camera,
  macOS Photo Booth, GNOME Cheese/Kamoso) and brings the photo or video back
  into the send view.
- **Events** — Attach > Event creates an event (name, description, start and
  end, place, "allow guests") and sends it as an event message.
- **Member labels** — a group member's label shows beside their name in a
  colored pill and stays current as they change it.
- **Select chats, by range** — shift-click picks a whole run of chats in the
  chat list, next to Archive, Mute, Pin, Mark read/unread and Delete.

### ⚙️ Settings

- **Auto-lock on idle** — the app lock can lock itself after a chosen time
  without use, measured from real system idle time.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.8.1...v1.9.0

## [v1.8.1] - 2026-10-19

- **WhatsApp links open here** — `wa.me`, `api.whatsapp.com/send` and
  `whatsapp://send` links open the number's chat with the link's text in
  the composer, from a browser or another app. Group invites
  (`chat.whatsapp.com`), channel links, business message links
  (`wa.me/message`) and contact QR links (`wa.me/qr`) open their chat,
  group or channel, or show the channel to follow first; a link that names
  no chat asks which one to send it to. On Windows the app registers
  itself as the handler of `whatsapp:` links and offers a bar above the
  chat list (and a row in Settings > General) until you pick it in
  Settings > Default apps.
- **Link previews load with a spinner** — a link's preview now shows a
  loading card while it's fetched, with a ✕ to cancel, instead of
  appearing a moment later.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.8.0...v1.8.1

## [v1.8.0] - 2026-10-17

### 🚀 Updates

- **Smaller downloads** — updating from one of the last five releases
  now downloads only what changed: a patch of a few percent of the exe,
  verified with checksums and its signature, then rebuilt from your
  running exe. Anything unexpected falls back to the full download.
- **Automatic updates** — with the toggle in Settings > Help on, the
  app checks a minute after it starts, then every few hours. A new
  version downloads and installs itself in the background, and the app
  restarts when no window is open (tray only); with a window open it
  offers to restart now, once. The toggle is on by default, so updates
  just happen; turn it off in Settings > Help.
- **Restart to update** — an automatic update doesn't interrupt: the
  installed version runs the next time the app starts, and the Help
  page's row restarts into it whenever you want.

### 🔧 Under the hood

- Update downloads are patch-aware end to end: patches are made from
  the last five releases' executables at release time, signed together
  with the checksums, and applied with a bspatch of the app's own. A
  release without a usable patch still updates through the full exe.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.7.0...v1.8.0

## [v1.7.0] - 2026-10-16

### 💬 Chats

- **React and Forward buttons** — hovering a message shows round buttons
  beside it: React opens the reaction bar, Forward forwards media,
  links, documents and stickers.
- **"Forwarded many times"** — messages forwarded around a lot say so,
  with a double arrow.
- **Waiting for this message** — a message that couldn't decrypt yet
  shows as that, and swaps in when it arrives. Its menu can Edit, React,
  Download and Report.
- **Select chats** — tick chats in the list to archive, mute, pin, mark
  read or unread, or delete them at once.
- **Report** — the message menu's Report asks first, and offers to block
  the sender.

### ⚙️ Settings

- **App lock** — lock the app behind a password: it asks when the app
  starts and after a chosen time without use. "Lock now" locks it at
  once; the password is kept only as a salted hash.
- **Search settings** — the Settings home is like WhatsApp's, with your
  profile on top and a search that finds the setting you name and jumps
  to it.
- **Video & voice** — pick the camera, microphone and speakers calls
  would use.
- **Storage** — choose which media downloads by itself. Photos can wait
  for the viewer, showing their thumbnail with a download button.
- **Profile pictures** — download every picture the window shows (as
  before), or only the open chat's and its members'.

### ✨ Emoji & links

- **Skin tones** — the emoji panel has a tone button, and a right click
  on an emoji offers it in every tone.
- **Chat links** — `whatsup://` links open the number's chat directly,
  and a `wa.me` link pasted into New chat opens it with its text filled
  in, without saving the contact.
- **Big emojis** — a message of one to three emojis shows them big,
  without a bubble.
- **Hover labels** — the icon-only buttons show what they do on a short
  hover.

### 🔧 Under the hood

- Contacts' abouts (with when they were set) show in the forward picker
  and contact info; Advanced chat privacy says whether it's on and stops
  Export chat; groups you left can only be deleted; the forward picker
  lists contacts under their own heading; the libra protocol was
  updated.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.6.0...v1.7.0

## [v1.6.0] - 2026-10-13

### 🔗 Chat links

- **Open chats from links** — `whatsup://` links open the number's
  chat directly, even when the app is already running. The scheme
  registers itself per user, no admin rights needed.
- **Paste wa.me links** — paste a `wa.me` or `api.whatsapp.com` link
  into New chat and hit Enter: the chat opens with its text filled in.
- **Nothing saved** — chatting from a link never adds the number to
  your contacts.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.5.0...v1.6.0

## [v1.5.0] - 2026-10-12

### ✅ Message info & groups

- **Group receipts, complete** — Message info in groups now shows every
  member's delivery and read, not just the few that arrived one by one.
- **Left groups stay honest** — a group you left or were removed from
  can't be sent to anymore: the composer says so, and its menu offers
  Delete group instead of Exit group.

### ✨ Sending

- **Big emojis** — a message of one to three emojis shows them big,
  without a bubble, like WhatsApp.
- **Captions keep your text** — attaching files turns the composer's
  text into the first caption, and closing the send view unsent keeps
  the shown caption. Slash commands run from captions too.
- **Bot list buttons** — a bot's list opens its sections and rows, and
  answering carries the row's ID, so bots in groups see it.

### 🔧 Under the hood

- Protocol library updated (socket, QR and receipt fixes); the README
  screenshots now render from the demo data with one command.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.4.1...v1.5.0

## [v1.4.1] - 2026-10-12

### 📚 Demo and docs

- **Fresh demo names** — the demo chats (`-demo`) use a new set of
  people, names and pictures, end to end: chats, senders, contacts,
  votes, mentions and the screenshot tool.
- **New tour screenshots** — the README tour and the landing page show
  the current UI, including the AFK list and Snippets pages.
- **No app changes** — chats, messages and settings work exactly as in
  v1.4.0; update only if you want the new demo and docs.

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.4.0...v1.4.1

## [v1.4.0] - 2026-10-10

### 💤 AFK list

- **Settings > AFK list** — its own page, next to Snippets: the away
  message (start AFK or update it without restarting the clock), who
  the reply goes to, and when it goes.
- **Who gets the reply** — add contacts by @mention, name or number,
  with suggestions as you type. Confine replies to them with the
  switch, or leave them for everyone; each sender still gets it once,
  marked ✓ on who did.
- **Reply hours** — replies go e.g. 21:00–07:00 (overnight works), or
  all day when the hours are empty.
- **Shorter reply** — the AFK reply is now just "💤 AFK: reason".

**Full Changelog**: https://github.com/latiefahmad/whatsupclients/compare/v1.3.0...v1.4.0

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
