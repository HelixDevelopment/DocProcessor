# AGENTS.md — Doc-Processor

## INHERITED FROM constitution/AGENTS.md

All rules in `constitution/AGENTS.md` (and the `constitution/Constitution.md` it references) apply unconditionally. This file's rules below extend them — they MUST NOT weaken any inherited rule. See parent root `CLAUDE.md` §6.AD for the Lava-specific incorporation context (29th §6.L cycle, 2026-05-14) and §6.AD-debt for the implementation-gap inventory. Use `constitution/find_constitution.sh` from the parent project root to resolve the absolute path of the submodule from any nested location.

## INHERITED FROM the Helix Constitution

This module is governed by the Helix Constitution. All rules in the
constitution's `AGENTS.md` and the `Constitution.md` it references apply
unconditionally. Locate the constitution from any nested depth via its
`find_constitution.sh` helper — do NOT hardcode a path (this module stays
fully decoupled and project-agnostic per §11.4.28).

Canonical reference: https://github.com/HelixDevelopment/HelixConstitution

## Module Identity

DocProcessor (`digital.vasic.docprocessor`, Go 1.25) is a standalone,
project-not-aware, fully decoupled documentation-processing module. It loads
project docs, builds structured feature maps, and tracks verification coverage
for QA automation. Provider-specific and project-specific behaviour is injected
at runtime via the `Translator`, `LLMAgent`, and `Config` contracts — the
module never imports a consuming project's namespace.

## Responsibilities

- Load and parse documentation (`pkg/loader`: Markdown, YAML, HTML, AsciiDoc,
  RST).
- Extract features and build a queryable `FeatureMap` (`pkg/feature`).
- Track per-feature verification coverage thread-safely (`pkg/coverage`).
- Build a directed inter-document link graph with JSON/Mermaid export
  (`pkg/docgraph`).
- Define the `LLMAgent` extraction contract and prompt templates (`pkg/llm`),
  with no hard dependency on any LLM provider.
- Externalise all user-facing strings (`pkg/i18n`) and load configuration
  (`pkg/config`).
- Expose a CLI: `docprocessor [--verbose|-v] <docs-directory>`.

## Testing Boundaries

- Build: `go build ./...`. Vet: `go vet ./...` (zero warnings).
- Tests: `go test ./... -race -count=1` covers unit, integration, stress,
  security, E2E, and automation suites.
- Integration/stress/security/E2E/automation tests run against real
  filesystems and the real CLI binary, not mocks; unit tests may use mocks.
- Agents MUST NOT weaken, stub, or skip tests, and MUST NOT introduce a
  dependency on any consuming project. Every change keeps build and vet green.
