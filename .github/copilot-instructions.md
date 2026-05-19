# Copilot instructions for `wendal/gor`

## Commands

- **Build:** `go build ./...`
- **Test all:** `go test ./...`
- **Single test:** `go test . -run TestName -count=1`  
  Example: `go test . -run TestCreatePostURL -count=1`
- **Vet:** `go vet ./...`

CI runs the build/test/vet matrix on Go 1.21, 1.22, and 1.23, so keep changes compatible with that range even though `go.mod` declares `go 1.24.13`.

## Big picture

- The root package `github.com/wendal/gor` holds the site compiler, content loaders, markdown/rendering helpers, widgets, plugins, and metadata types.
- The CLI lives in `./gor` (`package main`) and dispatches commands like `config`, `new/init`, `posts`, `payload`, `compile/build/c`, `post`, `addimg`, `http`, `preview`, and `pprof`.
- Content follows the Ruhoh-style layout: `config.yml`, `site.yml`, `pages/`, `posts/`, `partials/`, `themes/<theme>/...`, `widgets/`, `media/`, and generated output under `compiled/`.
- `Compile()` is the main pipeline: build payload, load partials and layouts, pre-render Mustache + Markdown, load widgets, copy resources, render pages/posts, run paginator logic, then execute plugins (`rss.xml`, `sitemap.xml`, `.nojekyll`).
- There are two payload paths: the legacy map-based flow (`BuildPayload`, `Mapper`) is the active compiler path; `payload_v2.go` is a struct-based variant that mirrors the same concepts.

## Conventions to preserve

- Front matter must start with `---`, end with `---`, and content follows after that delimiter.
- Posts need a `title`; `date` can be inferred from the filename prefix `YYYY-MM-DD-...` when omitted.
- `CreatePostURL` is the canonical permalink formatter. It expands `:year`, `:month`, `:day`, `:title`, `:filename`, and `:categories`.
- `EncodePathInfo` / `DecodePathInfo` define the URL-safe path encoding used for tags and categories.
- Layout rendering is recursive: a layout may declare its own `layout`, and `RenderInLayout` will wrap again.
- Markdown is rendered with Blackfriday; `{:toc}` enables TOC insertion.
- Use `Mapper`, `ToStruct`, `AsStrings`, and `ReadYmlReader` for dynamic config/content coercion. `ReadYmlReader` accepts JSON as well as YAML.
- Keep the Windows filename sanitization in `post.go` intact; new post titles must avoid reserved device names and invalid characters.
- Use the existing debug helpers (`D`, `PrintJson`) rather than adding ad hoc logging patterns.
