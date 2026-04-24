# CLAUDE.md

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

## Project Overview

DocProcessor is a Go module for loading project documentation, building structured feature maps, and tracking verification coverage. Part of the HelixQA ecosystem.

**Go module:** `digital.vasic.docprocessor`

## Build Commands

```bash
go build ./...                        # Build
go test ./... -race -count=1          # Test with race detection
go vet ./...                          # Static analysis
make all                              # tidy + vet + test + build
make test-cover                       # Test with coverage report
```

## MANDATORY Rules

- **NO test may ever be removed, disabled, skipped, or left broken**
- All tests must pass with `go test ./... -race -count=1`
- All source files must have SPDX license headers (Apache-2.0)
- CoverageTracker MUST remain thread-safe (sync.RWMutex)
- DocGraph MUST remain thread-safe (sync.RWMutex)
- LLMAgent interface MUST NOT have module-level dependencies

## Package Layout

```
pkg/loader/    - Loader interface, Document, Section, markdown/yaml parsers, scanner
pkg/feature/   - Feature, FeatureMap, FeatureMapBuilder, categories, screens
pkg/coverage/  - CoverageTracker (thread-safe), CoverageReport, Evidence, Issue
pkg/docgraph/  - DocGraph, Node, Edge, JSON/Mermaid export
pkg/llm/       - LLMAgent interface, RawFeature, prompt templates
pkg/config/    - Config from .env files
cmd/docprocessor/ - CLI entry point
```

## Test Types

- Unit tests (`*_test.go`)
- Integration tests (`*_integration_test.go`)
- Stress tests (`*_stress_test.go`) -- concurrent operations
- Security tests (`*_security_test.go`) -- path traversal, large files, malformed input
- E2E tests (`e2e_test.go`) -- full pipeline with mock LLMAgent
- Automation tests (`automation_test.go`) -- build validation, package structure

## Key Patterns

- `CoverageTracker`: Read operations use `RLock()`, write operations use `Lock()`
- `DocGraph`: Thread-safe with `sync.RWMutex`
- Feature IDs: Deterministic via `GenerateID(name)` -- slug + hash suffix for long names
- MaxFileSize: 10 MB limit on loaded files
- LLMAgent: Injected interface, no module-level dependency on LLMOrchestrator

## Dependencies

- `github.com/stretchr/testify` (testing)
- `gopkg.in/yaml.v3` (YAML parsing)

## Definition of Done

A change is NOT done because code compiles and tests pass. "Done" requires pasted
terminal output from a real run of the real system, produced in the same session as
the change. Coverage and passing suites measure the LLM's model of the product, not
the product.

1. **No self-certification.** *Verified, tested, working, complete, fixed, passing*
   are forbidden in commits, PRs, and agent replies without accompanying pasted
   output from a same-session real-system run.
2. **Demo before code.** Every task begins with the runnable acceptance demo below.
3. **Real system.** Demos run against real artifacts — built binaries, live
   databases, instrumented devices — not mocks/stubs/in-memory fakes.
4. **Skips are loud.** `t.Skip` / `@Ignore` / `xit` / `it.skip` without a trailing
   `SKIP-OK: #<ticket>` annotation fails `make ci-validate-all`.
5. **Contract tests on every seam.** Any change touching a module↔module boundary
   runs one roundtrip test asserting the wire format on both sides.
6. **Evidence in the PR.** PR body contains a fenced `## Demo` block with exact
   command(s) + output.

### Acceptance demo for this module

```bash
# TODO — replace with a 10-line real-system demo. See examples in
# HelixAgent/docs/development/dod-dropin/templates/CLAUDE_md_clause.md
```
