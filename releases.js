// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.14",
  "date": "2026-11-06",
  "items": [
    "**The snapshot diagnostic reads the blob reference** — v1.9.13 tried to parse the snapshot itself and failed on every answer, because a snapshot arrives as a reference to a blob fetched separately. The log now records that reference (handle, file size, hash) and the patches that follow it, and says plainly when an answer carries no snapshot at all."
  ]
};
