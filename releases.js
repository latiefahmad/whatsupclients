// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.9.12",
  "date": "2026-11-06",
  "items": [
    "**A settings channel that never synced is fetched again on connect** — a channel with no local state was skipped when the reconnect ran, because the pass only looked at channels it had marked done. WhatsApp can't apply changes to a channel the computer holds no state for, so those changes (a message deleted for me, a cleared chat) stayed queued and never arrived. It's fetched again now, and the queued changes go out once it has a state."
  ]
};
