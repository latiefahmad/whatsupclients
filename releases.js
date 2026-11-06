// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.11",
  "date": "2026-11-06",
  "items": [
    "**One broken settings channel no longer blocks the others** — when a channel failed to sync, the reconnect gave up and skipped every channel after it, so their changes (a message deleted for me, a cleared chat) never arrived either. Each one is tried on its own now.",
    "**A change is no longer sent into a channel with no local state** — a channel this computer holds no state for can't take a change: WhatsApp answers every send 409 and can't apply its conflicting patches either, so the change was lost and the send never had a chance. It's kept and sent once the channel has a state again, instead of being pushed into that."
  ]
};
