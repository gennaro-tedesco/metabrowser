## What this project is

`metabrowser` is a local epub library browser. Given a directory of epub files (flat or nested to any depth), it scans all epubs in parallel, extracts metadata from each file's internal OPF, and serves a browser UI on `localhost:7070`.

It has no database, no persistent cache, and no external runtime dependencies. The entire application ships as a single Go binary with the frontend embedded.

______________________________________________________________________

## How to build and run

```bash
go build -o metadata .
./metadata /path/to/epub/directory
# optional: override address
./metadata /path/to/epub/directory localhost:8080
```

Then open `http://localhost:7070` in a browser.

**Critical:** `static/` is embedded into the binary at compile time via `//go:embed static` in `main.go`. Editing files in `static/` has no effect until you rebuild. Always rebuild after any change to Go or static files.

______________________________________________________________________

## Project structure

```
.
├── main.go               # Entry point: validates CLI arg, calls Scan, calls Serve
├── go.mod                # Module: metabrowser, go 1.21, no external dependencies
├── model/
│   └── book.go           # Book and Library structs (JSON-serialisable)
├── scanner/
│   ├── epub.go           # Opens epub as ZIP, locates OPF, parses Dublin Core metadata
│   └── scanner.go        # Walks directory tree, worker pool, builds Library index
├── server/
│   └── server.go         # net/http: /api/library, /cover/, embedded static files
└── static/
    ├── index.html        # Single-page shell: sidebar + list + modal DOM skeleton
    ├── theme.css         # ONLY Catppuccin Mocha palette + font variables (:root)
    ├── style.css         # All layout and component CSS; imports theme.css
    └── app.js            # All frontend logic (vanilla JS, no framework)
```

______________________________________________________________________

## Architecture

### Data flow

```
startup
  └─ scanner.Scan(root)
       ├─ filepath.WalkDir → collect all .epub paths
       ├─ worker pool (NumCPU goroutines)
       │    └─ each worker: scanner.parseBook → scanner.parseEpub
       │         └─ open epub as zip/archive
       │              ├─ read META-INF/container.xml → get OPF path
       │              └─ parse OPF: dc:title, dc:creator, dc:language,
       │                            dc:subject, <meta name="calibre:series">
       └─ collect []Book, sort by title, build group index

HTTP GET /api/library
  └─ returns full Library as JSON (computed once at startup, never updated)

browser
  └─ fetch /api/library once on load
  └─ render sidebar groups from lib.groups
  └─ render book list from lib.books
  └─ all filtering is client-side (no round-trips after initial load)
```

______________________________________________________________________

## Style rules

- No comments in Go or JS code
- No external Go dependencies — stdlib only
- No JS framework, no build toolchain
- CSS variables only — no hardcoded colours or font names in `style.css`
- All colour/font changes go in `theme.css` exclusively
- Minimise diffs: prefer targeted edits over full rewrites
