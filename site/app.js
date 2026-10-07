// Detect the visitor's OS: highlight the matching download card and point
// the hero button at it. Links use releases/latest/download/, so they
// always serve the newest published release without any API calls.
(function () {
  var ua = navigator.userAgent || "";
  var os = /Win/i.test(navigator.platform || ua) ? "windows"
    : /Linux/i.test(navigator.platform || ua) ? "linux"
    : null;

  var cta = document.getElementById("cta-main");
  var note = document.getElementById("cta-note");

  if (!os) return;

  var cards = document.querySelectorAll('.dl-card[data-os="' + os + '"]');
  if (!cards.length) return;

  // Prefer the installer on Windows, the single binary on Linux.
  var pick = cards[0];
  cards.forEach(function (c) { c.classList.add("recommended"); });

  var link = pick.querySelector("a.btn");
  if (link && cta) {
    cta.href = link.href;
    cta.textContent = "";
    cta.innerHTML = '<svg class="ic"><use href="#i-download"/></svg>Download untuk '
      + (os === "windows" ? "Windows" : "Linux");
  }
  if (note) {
    note.textContent = os === "windows"
      ? "Terdeteksi Windows — tombol di atas mengunduh installer terbaru."
      : "Terdeteksi Linux — tombol di atas mengunduh binary terbaru.";
  }
})();

// Theme toggle: stored choice wins, applied pre-paint by the head script.
(function () {
  var btn = document.getElementById("theme-toggle");
  if (!btn) return;
  btn.addEventListener("click", function () {
    var next = document.documentElement.getAttribute("data-theme") === "light" ? "dark" : "light";
    document.documentElement.setAttribute("data-theme", next);
    try { localStorage.setItem("wuc-theme", next); } catch (e) {}
  });
})();

// Mobile menu: toggle the dropdown, close it when a link is picked.
(function () {
  var btn = document.getElementById("menu-btn");
  var links = document.getElementById("nav-links");
  if (!btn || !links) return;
  function close() {
    links.classList.remove("open");
    btn.setAttribute("aria-expanded", "false");
    btn.innerHTML = '<svg class="ic"><use href="#i-menu"/></svg>';
  }
  btn.addEventListener("click", function () {
    var open = links.classList.toggle("open");
    btn.setAttribute("aria-expanded", String(open));
    btn.innerHTML = '<svg class="ic"><use href="#' + (open ? "i-close" : "i-menu") + '"/></svg>';
  });
  links.addEventListener("click", function (e) {
    if (e.target.closest("a")) close();
  });
})();

// Scroll progress bar + back-to-top button + active nav link.
(function () {
  var bar = document.getElementById("progress-bar");
  var top = document.getElementById("to-top");
  function onScroll() {
    var max = document.documentElement.scrollHeight - window.innerHeight;
    if (bar) bar.style.width = (max > 0 ? (window.scrollY / max) * 100 : 0) + "%";
    if (top) top.classList.toggle("show", window.scrollY > window.innerHeight * 0.8);
  }
  window.addEventListener("scroll", onScroll, { passive: true });
  window.addEventListener("resize", onScroll);
  if (top) top.addEventListener("click", function () {
    window.scrollTo({ top: 0, behavior: "smooth" });
  });
  onScroll();

  var sections = document.querySelectorAll("main section[id]");
  var navA = document.querySelectorAll('#nav-links a[href^="#"]');
  if ("IntersectionObserver" in window && sections.length) {
    var spy = new IntersectionObserver(function (entries) {
      entries.forEach(function (en) {
        if (!en.isIntersecting) return;
        navA.forEach(function (a) {
          a.classList.toggle("active", a.getAttribute("href") === "#" + en.target.id);
        });
      });
    }, { rootMargin: "-40% 0px -55% 0px" });
    sections.forEach(function (s) { spy.observe(s); });
  }
})();

// Scroll reveal for cards, screenshots and stats.
(function () {
  var els = document.querySelectorAll(".card, .shot-row figure, .hero-shot, .stat");
  if (!("IntersectionObserver" in window) || !els.length) return;
  els.forEach(function (el) { el.classList.add("reveal"); });
  var io = new IntersectionObserver(function (entries) {
    entries.forEach(function (en) {
      if (en.isIntersecting) {
        en.target.classList.add("in");
        io.unobserve(en.target);
      }
    });
  }, { threshold: 0.12 });
  els.forEach(function (el) { io.observe(el); });
})();

// Release info: version badge + the first highlights of the newest section.
// Primary source is site/releases.js (generated from CHANGELOG.md) — it
// works with a private repo, offline, and without any rate limit. The raw
// GitHub fetch is only a fallback when that file is missing.
(function () {
  function esc(s) {
    return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
  function bold(s) {
    // escape first, then turn **x** into <b>x</b>
    return esc(s).replace(/\*\*(.+?)\*\*/g, "<b>$1</b>");
  }

  function render(data) {
    var tag = document.getElementById("ver-tag");
    if (tag && data.ver) {
      tag.textContent = data.ver;
      tag.href = "https://github.com/latiefahmad/whatsupclients/releases/tag/" + data.ver;
    }
    var newsTag = document.getElementById("news-tag");
    if (newsTag && data.ver) newsTag.textContent = data.ver;
    var body = document.getElementById("news-body");
    if (!body) return;
    var items = (data.items || []).slice(0, 4);
    body.innerHTML = items.length
      ? (data.ver ? "<h4>" + esc(data.ver) + (data.date ? " &middot; " + esc(data.date) : "") + "</h4>" : "")
        + "<ul><li>" + items.map(bold).join("</li><li>") + "</li></ul>"
      : '<p class="muted">Catatan rilis lengkap ada di halaman GitHub.</p>';
  }

  // 1) the generated snapshot shipped with the page
  if (window.WUC_RELEASE && window.WUC_RELEASE.ver) {
    render(window.WUC_RELEASE);
    return;
  }

  // 2) fallback: read CHANGELOG.md from GitHub (public repos only)
  function fail() {
    var body = document.getElementById("news-body");
    if (body) {
      body.innerHTML = '<p class="muted">Tidak bisa memuat catatan rilis. '
        + '<a href="https://github.com/latiefahmad/whatsupclients/releases/latest" '
        + 'target="_blank" rel="noopener">Buka halaman rilis &rarr;</a></p>';
    }
  }
  function parse(md) {
    var lines = md.split("\n");
    var ver = null, items = [], inSection = false, inItem = false;
    for (var i = 0; i < lines.length; i++) {
      var t = lines[i].trim();
      var head = t.match(/^##\s*\[?(v[\w.-]+)\]?/);
      if (head) {
        if (inSection) break;
        ver = head[1];
        inSection = true;
        inItem = false;
        continue;
      }
      if (!inSection) continue;
      if (!t) { inItem = false; continue; }
      if (t.charAt(0) === "#") { inItem = false; continue; }
      if (t.indexOf("- ") === 0) { items.push(t.slice(2)); inItem = true; }
      else if (inItem) { items[items.length - 1] += " " + t; }
    }
    return { ver: ver, items: items };
  }
  fetch("https://raw.githubusercontent.com/latiefahmad/whatsupclients/main/CHANGELOG.md", { cache: "no-cache" })
    .then(function (r) { if (!r.ok) throw new Error("http " + r.status); return r.text(); })
    .then(function (md) {
      var data = parse(md);
      if (!data.ver) throw new Error("no section");
      render(data);
    })
    .catch(fail);
})();
