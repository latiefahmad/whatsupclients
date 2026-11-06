// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.15",
  "date": "2026-11-06",
  "items": [
    "**A refused snapshot now reports whether its MAC can be reproduced** — the log lines from v1.9.14 showed the snapshot is a 495 KB blob with 26 patches, so it isn't missing or truncated. The diagnostic now downloads that blob, counts its records, recomputes the state hash over exactly those records, and prints the MAC it derives next to the one the snapshot carries. A `match=true` means the hash was rebuilt correctly and the key differs; a `match=false` points at the records themselves."
  ]
};
