// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.19",
  "date": "2026-11-12",
  "items": [
    "**Messages no longer get stuck on networks that slow down port 443** — WhatsApp Web opens two connection endpoints at once on every connect and keeps whichever answers first, because some networks throttle or block long-lived connections on 443. The app now does the same, so delivery keeps working where it used to stall until a manual reconnect.",
    "**Every settings sync names the version it starts from** — a change that silently fails to reach your phone used to look like one that succeeded in the log; now each send records where it started and where it ended, without needing debug mode."
  ]
};
