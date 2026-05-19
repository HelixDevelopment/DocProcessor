# DocProcessor

Documentation processing and feature map extraction for QA automation.

[![License](https://img.shields.io/badge/license-Apache%202.0-blue)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.25%2B-00ADD8)](go.mod)
[![Anti-Bluff](https://img.shields.io/badge/anti--bluff-CONST--035-red)](CONSTITUTION.md)
[![i18n](https://img.shields.io/badge/i18n-CONST--046-green)](pkg/i18n/translator.go)

## Overview

DocProcessor is a **standalone**, **project-not-aware**, **fully decoupled** Go
module (per CONST-051(B)) that loads project documentation, builds structured
feature maps, and tracks verification coverage. It is designed to work with LLM
agents for intelligent feature extraction, but also includes heuristic-based
extraction for offline use.

**Round 220 (2026-05-19) deep-doc + test-matrix enrichment.** This README
matches actual capability — every claim below is exercised by an automated
test or Challenge script in this repository (per CONST-048 invariants 1, 3, 5,
6 and the §11.4 anti-bluff covenant).

> Verbatim 2026-05-19 operator mandate (CONST-049 §11.4.17):
> *"all existing tests and Challenges do work in anti-bluff manner - they
> MUST confirm that all tested codebase really works as expected! We had
> been in position that all tests do execute with success and all
> Challenges as well, but in reality the most of the features does not
> work and can't be used! This MUST NOT be the case and execution of tests
> and Challenges MUST guarantee the quality, the completition and full
> usability by end users of the product!"*

## Quick Start

```bash
# Clone (per CONST-056: install upstreams immediately if present)
git clone git@github.com:HelixDevelopment/DocProcessor.git
cd DocProcessor
install_upstreams                  # if Upstreams/ recipe dir present

# Build
go build ./...

# Run unit + integration + stress + security + E2E + automation tests
go test ./... -race -count=1

# Build the CLI
go build -o bin/docprocessor ./cmd/docprocessor

# Run against a docs directory
./bin/docprocessor /path/to/docs              # terse summary
./bin/docprocessor --verbose /path/to/docs    # per-feature/screen/workflow lines
```

## Architecture

DocProcessor is organised into seven packages — six domain packages plus
`pkg/i18n` for the CONST-046 string-externalisation contract:

| Package          | Purpose                                                                                |
|------------------|----------------------------------------------------------------------------------------|
| `pkg/loader`     | Document loading + parsing (Markdown, YAML, HTML, AsciiDoc, RST)                       |
| `pkg/feature`    | Feature extraction, FeatureMap building, FeatureMapBuilder                             |
| `pkg/coverage`   | Thread-safe coverage tracking with RWMutex                                             |
| `pkg/docgraph`   | Inter-document link graph with JSON/Mermaid export                                     |
| `pkg/llm`        | LLMAgent interface + prompt templates (no hard dependency on any provider)             |
| `pkg/config`     | Configuration loading from `.env` files                                                |
| `pkg/i18n`       | Translator contract + NoopTranslator default (CONST-046 no-hardcoded-content)          |

### Processing Pipeline

```
Load Docs -> Parse Sections -> Extract Features -> Build FeatureMap -> Enrich (LLM) -> Track Coverage
```

1. **Load & Parse** — scan project tree for documentation files in configured formats.
2. **Extract Features** — heuristic extraction (offline) or LLM-powered extraction.
3. **Build Feature Map** — structured, queryable map with categories + platform matrix.
4. **Enrich** — optional `LLMAgent` infers screens and generates test steps.
5. **Track Coverage** — thread-safe per-platform verification tracking.

## CLI surface (round 209 grammar)

```
docprocessor [--verbose|-v] <docs-directory>
```

| Flag         | Effect                                                                       |
|--------------|------------------------------------------------------------------------------|
| `--verbose`  | Emit per-feature / per-screen / per-workflow lines after the summary block.  |
| `-v`         | Alias for `--verbose`.                                                       |
| (no flag)    | Terse summary: format-list, loaded-count, summary-header, feature-map/      |
|              | doc-graph counts, per-category + per-platform lines, completion line.       |

### Wire-output message IDs (CONST-046)

Every user-facing line is emitted via the injected `pkg/i18n.Translator`, never
as a hardcoded English literal. The full message-ID catalogue:

| Round | Message ID                                  | Emitted when                                   |
|-------|---------------------------------------------|------------------------------------------------|
| 97    | `docprocessor_cli_usage`                    | argv < 2 — usage line                          |
| 97    | `docprocessor_cli_error_loading_docs`       | loader returns an error                        |
| 97    | `docprocessor_cli_loaded_documents`         | loader returns N documents                     |
| 97    | `docprocessor_cli_error_building_feature_map` | feature builder returns an error             |
| 97    | `docprocessor_cli_feature_map_summary`      | summary block — F / S / W counts               |
| 97    | `docprocessor_cli_doc_graph_summary`        | summary block — nodes / edges                  |
| 97    | `docprocessor_cli_category_line`            | per-category count line                        |
| 97    | `docprocessor_cli_platform_line`            | per-platform count line                        |
| 209   | `docprocessor_cli_help_header`              | argv < 2 — startup banner second line          |
| 209   | `docprocessor_cli_path_invalid`             | empty / whitespace-only docs-directory arg     |
| 209   | `docprocessor_cli_error_resolving_path`     | `filepath.Abs` failure                         |
| 209   | `docprocessor_cli_no_docs_found`            | loader returns 0 documents (was silent pre-209)|
| 209   | `docprocessor_cli_format_summary`           | supported-formats line at run-start            |
| 209   | `docprocessor_cli_summary_header`           | summary section heading                        |
| 209   | `docprocessor_cli_feature_line`             | per-feature line (verbose mode only)           |
| 209   | `docprocessor_cli_screen_line`              | per-screen line (verbose mode only)            |
| 209   | `docprocessor_cli_workflow_line`            | per-workflow line (verbose mode only)          |
| 209   | `docprocessor_cli_done`                     | completion line with elapsed ms                |

All 18 IDs are present in `pkg/i18n/bundles/active.en.yaml` and asserted by
`TestRunCLI_BundleContainsAllRound209MsgIDs` in `cmd/docprocessor/main_test.go`.

## Key Interfaces

- `loader.Loader` — load documents from filesystem
- `feature.FeatureMapBuilder` — build feature maps from documents
- `coverage.CoverageTracker` — track feature verification status
- `llm.LLMAgent` — injected LLM for intelligent extraction (no hard dependency)
- `i18n.Translator` — externalised user-facing strings (CONST-046)

## Configuration

Copy `.env.example` to `.env` and customise:

```bash
HELIX_DOCS_ROOT=./docs
HELIX_DOCS_AUTO_DISCOVER=true
HELIX_DOCS_FORMATS=md,yaml,html,adoc,rst
```

> **CONST-053 reminder.** `.env` MUST be `chmod 600` and is git-ignored.
> Only `.env.example` (placeholder values) is committed.

## Testing

```bash
make test          # all tests
make test-race     # all tests with race detection
make test-cover    # all tests with coverage report (coverage.html)
```

**Test-type coverage (per CONST-050(B) 100%-test-type-coverage mandate):**

| Test type    | File(s)                                                                                  | Real-infra? |
|--------------|------------------------------------------------------------------------------------------|-------------|
| Unit         | `pkg/**/*_test.go` (excluding suffix-tagged), `cmd/docprocessor/main_test.go`            | n/a (mocks OK) |
| Integration  | `pkg/loader/loader_integration_test.go`                                                  | real fs     |
| Stress       | `pkg/coverage/tracker_stress_test.go`, `pkg/docgraph/graph_stress_test.go`,             | real         |
|              | `pkg/loader/loader_stress_test.go`                                                       |              |
| Security     | `pkg/config/config_security_test.go`, `pkg/loader/loader_security_test.go`,             | real         |
|              | `security_test.go`                                                                        |              |
| E2E          | `e2e_test.go`                                                                            | real binary  |
| Automation   | `automation_test.go`                                                                     | real binary  |
| Challenge    | `challenges/scripts/docprocessor_cli_challenge.sh` (round 220) + 8 other Challenge scripts | real binary |

Full per-test-file coverage matrix in [`docs/test-coverage.md`](docs/test-coverage.md).

## Challenge scripts

```bash
bash challenges/scripts/docprocessor_cli_challenge.sh    # round 220 — CLI end-to-end + paired mutation
bash challenges/scripts/chaos_failure_injection_challenge.sh
bash challenges/scripts/ddos_health_flood_challenge.sh
bash challenges/scripts/host_no_auto_suspend_challenge.sh
bash challenges/scripts/no_suspend_calls_challenge.sh
bash challenges/scripts/scaling_horizontal_challenge.sh
bash challenges/scripts/stress_sustained_load_challenge.sh
bash challenges/scripts/ui_terminal_interaction_challenge.sh
bash challenges/scripts/ux_end_to_end_flow_challenge.sh
```

Every Challenge captures positive runtime evidence per CONST-035 / Article XI
§11.9 and the §11.4 anti-bluff covenant. A passing Challenge is a claim that
the feature works end-to-end for an end user, not merely that the binary
exited 0.

## Governance

- [`CONSTITUTION.md`](CONSTITUTION.md) — module-specific tightenings on top of the canonical root.
- [`CLAUDE.md`](CLAUDE.md) — AI-agent operating manual (cascaded from constitution submodule per CONST-049).
- [`AGENTS.md`](AGENTS.md) — generic agent manual peer of CLAUDE.md.
- Canonical root: [HelixConstitution](https://github.com/HelixDevelopment/HelixConstitution).

## Decoupling guarantee (CONST-051(B))

DocProcessor does NOT import any consuming-project namespace. The module ID is
`digital.vasic.docprocessor`; no path under `pkg/**` or `cmd/**` references a
specific consumer. Project-specific behaviour is injected at runtime via the
`Translator`, `LLMAgent`, and `Config` contracts.

## License

Apache License 2.0. See [LICENSE](LICENSE).
