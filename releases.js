// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.16",
  "date": "2026-11-06",
  "items": [
    "**The snapshot diagnostic now rules out the key and the version** — v1.9.15 showed the blob is complete (495277 bytes, 2260 records, no short values) yet its MAC still doesn't reproduce. The log now tries the MAC against each of the key's five subkeys and against the versions either side of the one the snapshot claims, and says which one matches if any does. No match means the records and the MAC themselves disagree, which no local key change fixes."
  ]
};
