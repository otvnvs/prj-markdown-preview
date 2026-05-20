## Markdown Preview Webapp

A live, side-by-side markdown previewer built with [go-app](https://go-app.dev) and WebAssembly. Type markdown on the left, see rendered HTML on the right — no page reloads, no JavaScript frameworks, just Go.

## Features

- Live preview as you type
- GitHub Flavored Markdown (GFM) via [goldmark](https://github.com/yuin/goldmark)
- Tables, task lists, strikethrough, autolinks
- Syntax-highlighted code blocks
- Clear button
- Responsive layout
- Dark mode support (follows system preference)

## Requirements

- Go 1.21+
- `make` (optional, for convenience targets)

## Quick Start

```sh
git clone https://github.com/otvnvs/prj-markdown-preview
cd prj-markdown-preview
go mod tidy

# Build the wasm binary
GOARCH=wasm GOOS=js go build -o docs/app.wasm .

# Run the dev server
go build -o markdown-previewer .
./markdown-previewer
# open http://localhost:8000
```

## Project Structure

```
.
├── main.go              # Component, server, and generate logic
├── go.mod
├── go.sum
├── web/
│   └── style.css        # Application styles
├── docs/                # Generated static site (GitHub Pages)
│   ├── index.html
│   ├── app.js
│   ├── app-worker.js
│   ├── app.wasm
│   ├── wasm_exec.js
│   └── web/
│       └── style.css
├── test.md              # Markdown test document
├── Makefile
└── README.md
```

## Makefile Targets

```sh
make wasm      # Build docs/app.wasm
make server    # Build ./markdown-previewer binary
make all       # Both of the above
make compress  # Build wasm then gzip + brotli compress it
make clean     # Remove build artefacts
```

## Deploying to GitHub Pages

The binary doubles as a static site generator via the `-generate` flag.

```sh
# 1. Generate index.html, app.js, app-worker.js, manifest, wasm_exec.js, web/
./markdown-previewer -generate -repo prj-markdown-preview

# 2. Build the wasm
GOARCH=wasm GOOS=js go build \
  -ldflags="-s -w" -trimpath \
  -o docs/app.wasm .

# 3. Commit and push
git add docs/
git commit -m "deploy"
git push
```

In your repository settings: **Pages → Source → Deploy from branch → `main` → `/docs`**.

The `-repo` flag sets the asset base path to `/prj-markdown-preview/` so that GitHub Pages serves everything from the correct subdirectory. Omit it when regenerating for a custom domain served from the root.

## Flags

| Flag        | Default | Description                                              |
|-------------|---------|----------------------------------------------------------|
| `-generate` | `false` | Generate static site instead of starting the server      |
| `-dir`      | `docs`  | Output directory for static generation                   |
| `-repo`     | `""`    | GitHub Pages repo name; sets asset base path via `app.GitHubPages()` |

## How It Works

### Markdown Rendering

Markdown is converted server-side (in the wasm) using goldmark with the GFM extension. The rendered HTML is written directly to the preview div's `innerHTML`:

```go
el := app.Window().GetElementByID("preview")
el.Set("innerHTML", m.renderMarkdown())
```

`app.Raw()` is intentionally avoided — it passes HTML through go-app's virtual DOM differ, which treats the entire block as a single node and truncates output after the first block element (e.g. only `<h1>` renders, dropping all subsequent `<p>`, `<ul>` etc). Writing `innerHTML` directly bypasses the differ entirely.

### Build Tag Split

The codebase uses a single `main.go` without build tags. `app.RunWhenOnBrowser()` is a no-op on the server and activates the wasm event loop in the browser — go-app's standard pattern for sharing one file across both environments.

### Static Generation

`app.GenerateStaticWebsite(dir, handler)` writes the HTML shell, JS loader, service worker, and PWA manifest. The `generate()` function additionally copies `wasm_exec.js` from `$GOROOT` and mirrors the `web/` directory into the output folder.

## WebAssembly Size

The compiled `.wasm` is approximately 14 MB uncompressed — this is the Go runtime bundled into every wasm binary.

| Technique                         | Approx size |
|-----------------------------------|-------------|
| Raw build                         | 14 MB       |
| `-ldflags="-s -w" -trimpath`      | 12 MB       |
| + `wasm-opt -Oz`                  | 9 MB        |
| + Brotli compression (transfer)   | ~2.5 MB     |
| + Gzip compression (transfer)     | ~3.5 MB     |

Brotli or gzip compression is the highest-impact step. go-app's handler does not compress the wasm automatically — serve pre-compressed files via a reverse proxy (nginx, Caddy) or GitHub Pages' built-in compression.

## Dependencies

| Package | Purpose |
|---------|---------|
| [go-app v11](https://github.com/maxence-charriere/go-app) | WebAssembly UI framework |
| [goldmark](https://github.com/yuin/goldmark) | CommonMark + GFM markdown parser |
| [google/uuid](https://github.com/google/uuid) | Required transitively by go-app |
