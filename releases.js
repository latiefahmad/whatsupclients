// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.13",
  "date": "2026-11-06",
  "items": [
    "**A refused app state snapshot now says what it held** — when WhatsApp answers a snapshot request, the log records the version it claims, the key it was sealed with, how many records it carries, and whether it came with patches after it. A refused snapshot used to leave only \"mismatching LTHash\", which can't tell a wrong key from an incomplete snapshot. Nothing changes when `-debug` isn't on."
  ]
};
