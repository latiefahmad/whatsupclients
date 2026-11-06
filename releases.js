// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.10",
  "date": "2026-11-06",
  "items": [
    "**App state sync diagnostics under `-debug`** — a settings channel whose hash stops verifying makes WhatsApp answer \"your phone is repairing sync\" and refuse changes, so this device and the phone drift apart. The log now records the version and hash before and after every send, every failed verify, and what a repair request will meet, so the next occurrence shows what breaks the hash instead of only that it broke. Nothing changes when `-debug` isn't on."
  ]
};
