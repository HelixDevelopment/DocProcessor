# DocProcessor — Pre-Integration Materials

**Revision:** 1
**Last modified:** 2026-07-15T11:21:01Z
**Purpose:** Consolidated pre-integration materials (gate before any integration/deployment work).

> This document consolidates and verifies the pre-integration materials for the
> DocProcessor Go module. It cross-references the existing `ARCHITECTURE.md`,
> `API_REFERENCE.md`, `README.md`, and `docs/` set rather than rewriting them.
> Every statement below is grounded in the real repository; items that cannot be
> determined from the repo are marked `UNKNOWN:` per the no-guessing rule
> (§11.4.6). The module is project-agnostic and MUST stay so (§11.4.28) — nothing
> here couples it to a consuming project.

---

## 1. Purpose / What it is

DocProcessor is a **standalone, project-not-aware, fully decoupled Go module**
(`go.mod:1` → `module digital.vasic.docprocessor`, Go 1.25 per `go.mod:3`) that
loads project documentation, builds structured feature maps, and tracks
verification coverage for QA automation. It supports LLM-driven feature
extraction via an injected agent AND heuristic (offline) extraction. Source of
truth: `README.md:10-16`, `CLAUDE.md` "Module Overview", `ARCHITECTURE.md`.

**HelixTranslate-family relationship (evidenced).** The module's git tag lineage
resolves to `helix_translate-2.3.1-10-g253f92e` (`git describe --tags`),
i.e. its most recent tag is `helix_translate-2.3.1`. This is the repo evidence
that DocProcessor is part of the HelixTranslate tag lineage / family. Beyond the
shared tag lineage, no source file references "HelixTranslate" — the relationship
is a release/tag-lineage relationship, not a code dependency.

- License: Apache-2.0 (`LICENSE`, `README.md:201-203`, SPDX headers in every `.go`).
- Decoupling guarantee (`README.md:194-199`): imports no consuming-project
  namespace; project-specific behaviour is injected at runtime via the
  `Translator`, `LLMAgent`, and `Config` contracts.

## 2. Architecture overview

Cross-reference `ARCHITECTURE.md` (component/sequence/class/state/flowchart
Mermaid diagrams) and `API_REFERENCE.md` (type + interface signatures). Summary
grounded in the real source tree:

**Packages** — seven `pkg/` packages plus the CLI under `cmd/docprocessor` and
one `internal/` package:

| Package | Purpose | Evidence |
|---|---|---|
| `pkg/loader` | Document loading + parsing | `pkg/loader/{loader,scanner,markdown,yaml_parser}.go`, `README.md:61` |
| `pkg/feature` | Feature extraction + `FeatureMap` building (`DefaultBuilder`) | `pkg/feature/{builder,convert,feature}.go`, `README.md:62` |
| `pkg/coverage` | Thread-safe (`sync.RWMutex`) coverage tracking | `pkg/coverage/tracker.go`, `ARCHITECTURE.md:173-179` |
| `pkg/docgraph` | Directed inter-document link graph, JSON/Mermaid export | `pkg/docgraph/graph.go`, `README.md:64` |
| `pkg/llm` | `LLMAgent` interface + prompt templates (no provider dependency) | `pkg/llm/{agent,prompts}.go`, `README.md:65` |
| `pkg/config` | Configuration loading from `.env` files / maps | `pkg/config/config.go`, `README.md:66` |
| `pkg/i18n` | `Translator` contract + `NoopTranslator` default (string externalisation) | `pkg/i18n/{translator,bundle}.go`, `README.md:67` |
| `cmd/docprocessor` | CLI entry point (`main` + `runCLI`) | `cmd/docprocessor/main.go` |
| `internal/archdoc` | Internal verifier that keeps `docs/ARCHITECTURE.md` factually consistent with the real source tree | `internal/archdoc/archdoc.go:2-5` (package doc) |

**Processing pipeline** (`README.md:69-79`, `ARCHITECTURE.md` flowchart
`:136-160`, and the CLI body `cmd/docprocessor/main.go:82-155`):

```
Load Docs → Parse Sections → Extract Features → Build FeatureMap → Enrich (LLM) → Track Coverage
```

1. **Load & Parse** — scan the docs tree for supported formats (`loader.LoadDir`).
2. **Extract Features** — heuristic (offline) or LLM-powered.
3. **Build Feature Map** — structured map with categories + platform matrix
   (`feature.NewBuilder(root).BuildFromDocs`, `main.go:100-101`).
4. **Enrich** — optional injected `LLMAgent` infers screens + test steps
   (`ARCHITECTURE.md` sequence diagram `:41-46`).
5. **Track Coverage** — thread-safe per-platform verification tracking.

**Formats it PROCESSES (input parsing, not export).** DocProcessor *reads and
parses* documentation; it does **not** itself export to PDF/DOCX. Default
supported input extensions: `md`, `yaml` (`yml` normalised to `yaml`), `html`,
`adoc`, `rst` — `pkg/config/config.go:26`, `.env.example` (`HELIX_DOCS_FORMATS`),
`pkg/loader/scanner.go`. `MaxFileSize = 10 MB` guards OOM
(`pkg/loader/scanner.go:14`, `ARCHITECTURE.md:187`).

> Note on the `.html`/`.pdf` siblings in this repo (e.g. `ARCHITECTURE.pdf`):
> these are exports of the repo's own documentation produced by an external
> pipeline. `UNKNOWN:` no `pandoc` / `weasyprint` / `wkhtmltopdf` / `libreoffice`
> reference exists anywhere in this module's scripts, Makefile, or source (grep
> returned zero matches), so the export tooling lives outside this module.

**Stack:** Go 1.25 standard library + `gopkg.in/yaml.v3` (only non-test runtime
dependency). Concurrency via `sync.RWMutex` in `pkg/coverage` and `pkg/docgraph`.

## 3. Dependencies

- **`.gitmodules`:** none present (no git submodules in this module).
- **`helix-deps.yaml`:** `deps: []` — declares **zero own-org submodule
  dependencies** (`helix-deps.yaml:16-17`). It is a leaf Go submodule; no
  `replace` directives point at own-org modules (`helix-deps.yaml:4-8`).
  `transitive_handling.recursive: true`, `conflict_resolution: operator-required`.
- **Go module dependencies (`go.mod:5-16`):**
  - Runtime: `gopkg.in/yaml.v3 v3.0.1`.
  - Test only: `github.com/stretchr/testify v1.11.1` (+ its indirect deps
    `davecgh/go-spew`, `kr/pretty`, `pmezard/go-difflib`, `rogpeppe/go-internal`,
    `gopkg.in/check.v1`).
- **System deps:** none required for the module itself beyond the Go 1.25
  toolchain (`Makefile`, `README.md:41-52`). No `pandoc`/`weasyprint`/database/
  broker/infra dependency — the module is pure-Go, filesystem-only.
- **Infra:** none. No database, no message broker, no external service. LLM
  usage is optional and injected via the `llm.LLMAgent` interface with **no hard
  dependency on any provider** (`ARCHITECTURE.md:183`, `README.md:65`).

## 4. Deploy / Distribution design

DocProcessor is distributed **as a Go module**, consumable two ways
(`README.md:41-52`, `CLAUDE.md` "Build & Test"):

1. **As a library** — importing `digital.vasic.docprocessor/pkg/{loader,feature,
   coverage,docgraph,llm,config,i18n}`. This is the primary integration surface
   (see the exported interfaces in `API_REFERENCE.md`). Consumers inject their own
   `Translator`, `LLMAgent`, and `Config`.
2. **As a CLI binary** — `go build -o bin/docprocessor ./cmd/docprocessor`
   (`README.md:47`), producing a single self-contained executable (i18n bundles
   are embedded via `//go:embed bundles/*.yaml`, so the binary is CWD-independent —
   `cmd/docprocessor/main.go:38-40`).

- **Container:** none. No `Dockerfile*` and no `compose*.yml` / `docker-compose*.yml`
  exist in this module (verified by directory listing). Distribution is
  source-module + optional compiled binary, not an image.
- **Build gates:** `Makefile` targets (`all: tidy vet test build`); tests run via
  `go test ./... -race -count=1` (unit / integration / stress / security / E2E /
  automation) plus Challenge scripts under `challenges/scripts/`
  (`README.md:168-185`).

## 5. Ports

`UNKNOWN: no listen port.` DocProcessor is a library + one-shot CLI, not a
server. Evidence: a source-wide grep for `net/http`, `ListenAndServe`, `http.Server`,
`.Serve(`, `bind`, `grpc`, and `:808x` in non-test `.go` files returned **zero
matches**; `cmd/docprocessor/main.go` runs `runCLI(...)` and exits
(`main.go:25-28`) with no network surface. There is no port to expose or map at
integration time.

## 6. Health

`UNKNOWN: no health endpoint.` There is no HTTP server, so no `/health`,
`/healthz`, or `/readyz` endpoint exists (same zero-match grep as §5). The
operational "health" signal for integration is instead the **process exit code**
of the CLI (`runCLI` returns an explicit exit code — `0` success, `1` on
usage/path/load/build errors: `main.go:59-155`) and the `go test ./...` /
Challenge-script pass/fail outcome (`README.md:144-185`).

## 7. How it boots

- **As a binary:** `docprocessor [--verbose|-v] <docs-directory>`
  (`README.md:81-92`, `cmd/docprocessor/main.go:59-171`). `main()` wires the
  production `Translator` (`i18n.NewBundleTranslator("en")`, falling back to
  `NoopTranslator` id-echo on bundle-load failure — `main.go:47-53`) and calls
  `runCLI`, which loads the docs dir, builds the feature map, prints a summary
  (feature/screen/workflow + doc-graph + per-category + per-platform counts), and
  exits. `--verbose`/`-v` adds per-feature/screen/workflow lines.
- **As a library:** it does not "boot" — it is consumed in-process. Entry points:
  `loader.NewDefaultLoader(formats)`, `feature.NewBuilder(root).BuildFromDocs(ctx,
  docs)`, `coverage.NewTracker()` (`API_REFERENCE.md`, `CLAUDE.md` "Key
  Interfaces"). No daemon, no background service.

## 8. Materials status (verify pass)

| Material | Present? | Grounded / evidence | Status |
|---|---|---|---|
| `README.md` | Yes | Overview, quick-start, package table, CLI grammar, test matrix | HAS-VERIFIED |
| `ARCHITECTURE.md` | Yes | Component/sequence/class/state/flowchart diagrams + design decisions; kept consistent by `internal/archdoc` | HAS-VERIFIED |
| `API_REFERENCE.md` | Yes | Type + interface signatures match source (`pkg/loader`, `pkg/feature`, …) | HAS-VERIFIED |
| `docs/test-coverage.md` | Yes | Per-test-file coverage matrix (7 test types + paired mutations) | HAS-VERIFIED |
| `docs/HOST_POWER_MANAGEMENT.md` | Yes | Host-safety companion doc + `scripts/host-power-management/` | HAS-VERIFIED (present; not integration-critical) |
| `go.mod` / `go.sum` | Yes | Module id, Go 1.25, dependency set | HAS-VERIFIED |
| `helix-deps.yaml` | Yes | Zero own-org deps declared (leaf module) | HAS-VERIFIED |
| `.env.example` | Yes | `HELIX_DOCS_ROOT` / `HELIX_DOCS_AUTO_DISCOVER` / `HELIX_DOCS_FORMATS` | HAS-VERIFIED |
| `Makefile` | Yes | build/test/test-race/test-cover/vet/fmt/tidy targets | HAS-VERIFIED |
| Ports doc | N/A | No listen port — library + CLI | HAS-VERIFIED (correctly absent, §5) |
| Health endpoint | N/A | No HTTP server — exit-code health, §6 | HAS-VERIFIED (correctly absent) |
| `Dockerfile` / `compose*.yml` | Absent | No container distribution (§4) | HAS-VERIFIED (correctly absent) |

**Open item to confirm at integration time (honest, not a blocker).**
`UNKNOWN:` authoritative upstream org. `README.md:36` quick-start clones
`git@github.com:HelixDevelopment/DocProcessor.git`, while `helix-deps.yaml:12`
states the authoritative remote is `vasic-digital/DocProcessor`. Both are
recorded in-repo; the definitive owning org should be confirmed against the live
git remote before any push/integration wiring (not determinable from static
files alone).

**Overall verdict: HAS-VERIFIED.** The pre-integration materials already exist,
are internally consistent, and are grounded in the real source tree. No content
was invented to fill a gap; the only `UNKNOWN:` items are genuine absences
(no port, no health endpoint, no container — because the module is a decoupled
Go library + one-shot CLI) plus the one upstream-org discrepancy noted above.
