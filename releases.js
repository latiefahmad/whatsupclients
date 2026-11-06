// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.18",
  "date": "2026-11-06",
  "items": [
    "**A collection that can't sync no longer grows without bound** — a collection whose snapshot never verifies can never accept a patch, so every delete and edit piled up for as long as the app ran and the queue was rewritten in full each time. Each collection now keeps at most 200 changes, dropping the oldest, and once it is full the app stops retrying it instead of repeating the same warning. Other collections are untouched."
  ]
};
