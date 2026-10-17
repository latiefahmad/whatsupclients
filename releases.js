// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.
// Re-run it whenever CHANGELOG.md gains a release section.
window.WUC_RELEASE = {
  "ver": "v1.8.0",
  "date": "2026-10-17",
  "items": [
    "**Smaller downloads** — updating from one of the last five releases now downloads only what changed: a patch of a few percent of the exe, verified with checksums and its signature, then rebuilt from your running exe. Anything unexpected falls back to the full download.",
    "**Automatic updates** — with the toggle in Settings > Help on, the app checks a minute after it starts, then every few hours. A new version downloads and installs itself in the background, and the app restarts when no window is open (tray only); with a window open it offers to restart now, once. The toggle is on by default, so updates just happen; turn it off in Settings > Help.",
    "**Restart to update** — an automatic update doesn't interrupt: the installed version runs the next time the app starts, and the Help page's row restarts into it whenever you want.",
    "Update downloads are patch-aware end to end: patches are made from the last five releases' executables at release time, signed together with the checksums, and applied with a bspatch of the app's own. A release without a usable patch still updates through the full exe."
  ]
};
