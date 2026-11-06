// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.17",
  "date": "2026-11-06",
  "items": [
    "**The snapshot blob is fetched once and analysed offline** — fetching it needs the media connection, which refreshes slowly and drops with the socket, so the diagnostic kept failing before it could report anything. The blob is now kept in the data directory and read from there on later attempts, so the analysis runs even when the media server can't be reached."
  ]
};
