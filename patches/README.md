# Dependency patches

`typesetting.patch` patches go-text (`github.com/go-text/typesetting`, Gio's
text library) to use far less memory for fonts, and `hypermeow.patch` patches
hypermeow (`github.com/polymorfa/hypermeow`) so it decrypts status updates.
`go.mod` replaces the modules with `third_party/typesetting` and
`third_party/hypermeow`, which `apply.sh` builds from the Go module cache and
which are not checked in:

```sh
sh patches/apply.sh
```

Run it after cloning, and again when a patch or a patched module's version in
`go.mod` changes. Until then the build fails with "replacement directory
./third_party/typesetting does not exist".

What `hypermeow.patch` does (marked `WhatsUpClients patch`):

- `message.go`: since November 2026 WhatsApp delivers status updates as
  `<status from="status@broadcast">` stanzas with the `<enc>` nodes a
  `<message>` had. hypermeow (and upstream whatsmeow) handled `<status>` as
  a plaintext newsletter update only, and acked and dropped the rest, so no
  status arrived. One with `<enc>` now goes through `handleEncryptedMessage`;
  the node keeps its tag, so the ack is still `class="status"`.
  `status_patch_test.go` checks it. Drop the patch once hypermeow does this.

After `go get github.com/polymorfa/hypermeow@main` run `apply.sh` again; if
the hunk no longer applies, regenerate the patch as below.

What `typesetting.patch` does (each change is marked `WhatsUpClients patch`):

- `font/opentype/tables/glyphs_glyf_src.go` and its users in `font/`: glyphs
  are parsed on demand. Upstream parses a font's whole `glyf` table when it
  loads, about ten times the table size in heap (17 MB for MS Gothic).
- `font/opentype/bytes.go`, `reader.go`: a `BytesReader` resource returns
  tables as subslices instead of copies (the bundled fonts, see
  `internal/ui/fonts.go`).
- `fontscan/mmap*.go`, `footprint.go`: system fonts are memory-mapped
  (copy-on-write) instead of read into the heap.

To move to another version of a patched module, change `go.mod`, run
`apply.sh`, fix any rejected hunks in `third_party/<module>`, and regenerate
the patch from a clean copy of the new version (typesetting shown):

```sh
cd "$(mktemp -d)" && cp -R "$(go env GOMODCACHE)/github.com/go-text/typesetting@<version>" ts &&
    chmod -R u+w ts && cd ts && git init -q && git add -A && git commit -qm upstream &&
    cp -R <repo>/third_party/typesetting/. . && git add -A &&
    git diff --cached > <repo>/patches/typesetting.patch
```
