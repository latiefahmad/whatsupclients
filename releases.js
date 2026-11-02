// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.8",
  "date": "2026-10-30",
  "items": [
    "**A bot no longer shows up as two chats** — a business bot sometimes opens its message under an ID this device was never told of, while the message itself says which number sent it. The same message was then filed under both, so the same bot appeared twice in the list — one entry with its name and one as \"Someone\" — and every message notified twice. Such a message now goes to the chat its number already had.",
    "**Chats already split this way are joined back** — the app folds them into the right chat once on first start, keeping each message a single time. On a real store this reunited 15 messages that had been duplicated."
  ]
};
