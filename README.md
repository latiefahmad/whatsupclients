# WhatsUp Clients

**WhatsApp on your desktop, as a native program — not a browser wearing its icon.**

[English](README.md) · [Bahasa Indonesia](README.id.md)

[![Build](https://github.com/latiefahmad/whatsupclients/actions/workflows/build.yml/badge.svg)](https://github.com/latiefahmad/whatsupclients/actions/workflows/build.yml)
[![Release](https://img.shields.io/github/v/release/latiefahmad/whatsupclients?sort=semver)](https://github.com/latiefahmad/whatsupclients/releases)
![Go](https://img.shields.io/github/go-mod/go-version/latiefahmad/whatsupclients)
![Platforms](https://img.shields.io/badge/platforms-Windows%20%7C%20Linux-25D366)

<img src="docs/screenshots/all-chat.png" alt="WhatsUp Clients showing all chats" width="900">

---

## What this is

Most desktop chat apps ship a browser engine and render one web page inside it. This one
doesn't. WhatsUp Clients draws every pixel itself on the GPU with [Gio](https://gioui.org)
and speaks WhatsApp's protocol directly through
[hypermeow](https://github.com/polymorfa/hypermeow).

That choice buys one thing above all: **the app stays live**. It runs in your tray, keeps
the connection open, and keeps everything in sync while you get on with your work — whether
or not a window is on screen.

## It stays in sync, for real

The part people notice first. Change something on your phone and it happens here, without
a reload:

- **Delete a message on your phone** → it disappears here.
- **Edit it there** → the edited text shows here.
- **Read it there** → the notification here goes away.
- **Clear a chat there** → the chat clears here.

WhatsApp moves these changes through private settings channels that linked devices fetch
and replay. If one of those channels gets stuck, the app used to keep believing it was
already up to date, and quietly stopped receiving *anything* from it — deletions included.
It now checks that a channel really stored something before marking it done, and reopens
one that didn't. When WhatsApp's own copy is unusable, it asks your phone for a fresh one,
exactly as WhatsApp's other linked devices do.

## It keeps working when you close it

- **Close the window and the app is still there.** The connection stays open in the tray,
  so messages arrive and notify you. The window's GPU memory is released, and comes back
  when you reopen it.
- **Scheduled messages send on time** — even with no window open, even with the app only
  in the tray. `/schedule 21:00 Invoice attached` and forget it.
- **AFK replies answer for you** while you're away, until you send something yourself.
- **Notifications arrive live**, with Reply and Mark as read on Windows.

## Screenshots

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/login.png" alt="Linking the app with a QR code"></td>
    <td width="50%"><img src="docs/screenshots/reactions.png" alt="Reacting to a message"></td>
  </tr>
  <tr>
    <td>Link with a QR code, like WhatsApp Web, and choose how much history to bring down.</td>
    <td>React, reply, forward, pin, star, select and delete — all from one right-click.</td>
  </tr>
  <tr>
    <td><img src="docs/screenshots/status.png" alt="Status updates"></td>
    <td><img src="docs/screenshots/msginfo.png" alt="Message info"></td>
  </tr>
  <tr>
    <td>View and post statuses: text, photos and videos, with audiences of your own.</td>
    <td>Message info shows who each of your messages reached, and when.</td>
  </tr>
  <tr>
    <td><img src="docs/screenshots/poll.png" alt="A poll and its votes"></td>
    <td><img src="docs/screenshots/communities.png" alt="Communities"></td>
  </tr>
  <tr>
    <td>Vote in a poll with one click, and see who picked what.</td>
    <td>Communities with their announcements and groups, and channels to follow.</td>
  </tr>
</table>

## What's built in

- Two windows into your chats: **all chats in one list**, or split with the conversation.
- **Drafts** that stay with each chat, **@mentions**, and bold / italic / strikethrough / code formatting.
- **Voice messages and video** played by the OS's own decoders — no codec bundled.
- **Pinned, archived and muted** chats, favourites, custom lists, disappearing messages, chat themes.
- **Albums and documents**, link previews with WhatsApp's big picture, group invite links.
- **View once** messages that open once, with screenshots blocked while one is on screen.
- **Several accounts**, with only the open one connected, to keep memory honest.
- **Light and dark themes**, WhatsApp's doodle wallpaper, System default that follows the OS.
- **Font size** from 70% to 200%, with Ctrl + `+` / `-` / `0`.
- **Updates from signed GitHub releases**, checked automatically and offered as *Update now* or *Later*.

Memory is the reason this project exists. Caches are capped by size, not entry count, and
after ten idle seconds the app hands memory back to the OS.

## Beyond what WhatsApp offers

Everything here is **off until you switch it on** in *Settings > Extra features*. Each one
only uses what WhatsApp already allows every linked device to do.

### Slash commands

Type `/` and a picker appears over the composer. Group admins get the management commands
without opening a single menu.

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/slash.png" alt="The slash command picker"></td>
    <td width="50%"><img src="docs/screenshots/calc.png" alt="The /calc command"></td>
  </tr>
  <tr>
    <td>The picker lists what the current chat allows, and takes Tab, arrows and Enter.</td>
    <td><code>/calc</code> works the sum out as you type, and picks up amounts from the message you reply to.</td>
  </tr>
</table>

| Command | What it does |
| --- | --- |
| `/add <contact>` | Adds people to the group |
| `/kick <member>` | Removes members |
| `/promote` / `/demote <member>` | Makes admins, or dismisses them |
| `/link` | Shows the group's invite link |
| `/lockdown on\|off` | Admins only, or everyone again |
| `/description <text>` | Changes the group description |
| `/sticker top#bottom` | Turns a photo into a sticker, meme text included |
| `/purge <count>` | Deletes your last messages for everyone |
| `/raffle <winners>` | Draws group members at random and announces them |
| `/calc <sum>` | `12 x 4500`, `15k x 3`, `15% x 80000` |
| `/schedule <when> <message>` | `21:00`, `2h`, `tomorrow 08:00`, `fri 18:00` |
| `/scheduled` | Lists what's queued in this chat: send now, or cancel |
| `/afk <reason>` | Replies for you until you send something; restrict it in *Settings > AFK list* |
| `/ghost` | No read receipts, shown offline, nothing sent — an ethically gray feature |
| `/snippet send <name>` | Sends a saved message |
| `/snippet save [name or text]` | Saves the message you reply to, or the text itself |
| `/catch` | Opens the stored payload of the message you reply to |

### Snippets

Save the messages you send over and over in *Settings > Snippets*, then send one with
`/snippet send`. The picker offers them as you type. Text can carry variables, filled in
when you send: `{name}`, `{first}`, `{mention}`, `{chat}`, `{me}`, `{quote}`, `{greeting}`,
`{date}`, `{time}` and `{day}`. Unknown `{...}` stays as it is; `\{name}` sends it literally.

### Privacy, for when someone's looking

<table>
  <tr>
    <td width="50%"><img src="docs/screenshots/privacy.png" alt="Privacy mode"></td>
    <td width="50%"><img src="docs/screenshots/afk-list.png" alt="AFK list and scheduled messages"></td>
  </tr>
  <tr>
    <td><b>Privacy mode</b>: names and messages become bars, pictures blur, all until you point at them. Ctrl+Shift+P, or the eye in the title bar. <b>Block screen recording</b> keeps the window out of captures.</td>
    <td><b>AFK list</b> limits who gets the automatic reply, and scheduled messages sit in the chat until their time comes.</td>
  </tr>
</table>

### Ethically gray features

A page of its own, each with a switch and its own warning:

- 🗑️ **Keep deleted messages** — messages and statuses deleted for everyone stay, marked Deleted.
- 👁️ **Replay view once** — open view once messages as often as you like.
- 👻 **`/ghost`** — read chats and statuses invisibly.

Turning one on opens a warning about privacy expectations. You must check the
acknowledgment and choose **Enable feature** before it takes effect. Turning it off is
immediate; turning it back on asks again.

## Download

Get the latest from [Releases](https://github.com/latiefahmad/whatsupclients/releases).
Prebuilt packages for **Windows** (amd64) and **Linux** (amd64) are attached to every release.

- `WhatsUpClients-Setup.exe` — installs for your user only. No admin rights.
- `WhatsUpClients-windows-amd64.exe` — the same app, no installer.
- `WhatsUpClients-linux-amd64` — the Linux binary.

To update, use *Settings > Help > Check for updates*. The app also checks on its own
shortly after start and then every four hours, and offers each release once. Every release
carries its [CHANGELOG](CHANGELOG.md) section as the release notes.

## Look around without an account

Demo mode starts the app with made-up chats and no network. Every screenshot above comes
from it.

```sh
go run ./cmd/whatsup -demo
```

## Build it yourself

You need the Go version in `go.mod`. On Linux, install Gio's
[system dependencies](https://gioui.org/doc/install/linux) first.

```sh
sh patches/apply.sh                 # once after cloning: builds the patched go-text
go run ./cmd/whatsup                # run it, and link with a QR code
go run ./cmd/whatsup -demo          # made-up chats, no network
go run ./cmd/whatsup -background    # start in the tray, no window
go run ./cmd/whatsup -debug         # log protocol traffic
go build ./...
```

Your data lives in `%AppData%\WhatsUpClients\whatsup.db` on Windows, and in the user config
directory elsewhere.

## How the pieces fit

```
cmd/whatsup/        the app
cmd/screenshot/     renders the UI to PNG, or beside a real WhatsApp screenshot
cmd/memprobe/       memory and frame-time benchmark (Windows)
internal/model/     what the UI sees: chats, messages, events, the Backend interface
internal/ui/        the Gio interface, from the chat list to the composer
internal/wa/        hypermeow backend and the SQLite message store
internal/command/   slash commands
internal/auto/      scheduled messages and AFK replies
internal/sticker/   makes stickers, with its own lossless WebP encoder
internal/video/     video and audio through the OS's decoders
internal/notify/    system notifications
internal/desktop/   tray, start at login, single instance
internal/update/    signed self-updates
internal/mock/      the demo backend
```

The UI never touches protocol types. `internal/wa` turns hypermeow events into
`internal/model` events, and the UI drains them on its own goroutine, so UI state needs no
locks.

## Contributing

Read [AGENTS.md](AGENTS.md) first. It covers the conventions and the Gio gotchas found the
hard way. Low memory is the point of this project, so measure with `cmd/memprobe` before
and after a change. Run `gofmt`, `go vet ./...` and `go build ./...` before sending it.

## License

[MIT](LICENSE).

## Disclaimer

Unofficial. Not affiliated with or endorsed by WhatsApp or Meta. Using a third-party client
may break WhatsApp's Terms of Service, so use it at your own risk.
