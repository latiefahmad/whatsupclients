// Generates releases.js from CHANGELOG.md (newest section).
// Run: node build-releases.js
// The landing page reads the generated file, so it works with a
// private repo, offline, and without any API rate limit.
"use strict";
const fs = require("fs");
const path = require("path");

const root = __dirname;
const md = fs.readFileSync(path.join(root, "CHANGELOG.md"), "utf8");

function parse(text) {
  const lines = text.split("\n");
  let ver = null, date = null, items = [], inSection = false, inItem = false;
  for (const raw of lines) {
    const t = raw.trim();
    const head = t.match(/^##\s*\[?(v[\w.-]+)\]?\s*(?:-\s*(\d{4}-\d{2}-\d{2}))?/);
    if (head) {
      if (inSection) break;
      ver = head[1];
      date = head[2] || null;
      inSection = true;
      inItem = false;
      continue;
    }
    if (!inSection) continue;
    if (!t) { inItem = false; continue; }
    if (t[0] === "#") { inItem = false; continue; }
    if (t.slice(0, 2) === "- ") { items.push(t.slice(2)); inItem = true; }
    else if (inItem) { items[items.length - 1] += " " + t; }
  }
  return { ver, date, items: items.slice(0, 4) };
}

const data = parse(md);
if (!data.ver) {
  console.error("build-releases: no `## [vX.Y.Z]` section found in CHANGELOG.md");
  process.exit(1);
}

const out = [
  "// Generated from CHANGELOG.md by build-releases.js — do not edit by hand.",
  "// Re-run it whenever CHANGELOG.md gains a release section.",
  "window.WUC_RELEASE = " + JSON.stringify(data, null, 2) + ";",
  "",
].join("\n");

fs.writeFileSync(path.join(root, "releases.js"), out, "utf8");
console.log("build-releases: wrote releases.js for " + data.ver
  + " (" + data.items.length + " items)");
