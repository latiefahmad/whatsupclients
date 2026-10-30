// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.7",
  "date": "2026-10-30",
  "items": [
    "**AI agent replies really show up now** — v1.9.6 read only one of the two shapes a bot's reply comes in, and newer bots send the other one: no submessages at all, one JSON document holding the text instead. Those replies still read \"This message couldn't load\". Both shapes are read now.",
    "**Bot replies that arrived before this are recovered** — they were stored as unreadable, but the message itself was kept, so the app reads them back once on first start and shows their text. Nothing else is touched."
  ]
};
