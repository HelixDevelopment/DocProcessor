# CLAUDE.md

## MANDATORY: Project-Agnostic / 100% Decoupled

**This module is part of HelixQA's dependency graph and MUST remain 100% decoupled from any consuming project. It is designed for generic use with ANY project, not just ATMOSphere.**

- **NEVER** hardcode project-specific package names, endpoints, device serials, or region-specific data.
- **NEVER** import anything from the consuming project.
- **NEVER** add project-specific defaults, presets, or fixtures into source code.
- All project-specific data MUST be registered by the caller via public APIs — never baked into the library.
- Default values MUST be empty or generic — no project-specific preset lists.

**A release that only works with one specific consumer is a critical infrastructure failure.** Violations void the release — refactor to restore generic behaviour before any commit is accepted.

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


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**

