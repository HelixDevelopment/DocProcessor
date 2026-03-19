# AGENTS.md

## MANDATORY: No CI/CD Pipelines

**NO GitHub Actions, GitLab CI/CD, or any automated pipeline may exist in this repository!**

- No `.github/workflows/` directory
- No `.gitlab-ci.yml` file
- No Jenkinsfile, .travis.yml, .circleci, or any other CI configuration
- All builds and tests are run manually or via Makefile targets
- This rule is permanent and non-negotiable

Instructions for AI coding agents working on DocProcessor.

## Project Context

DocProcessor is a Go module (`digital.vasic.docprocessor`) that processes project documentation into structured feature maps for QA automation. It is part of the HelixQA ecosystem alongside LLMOrchestrator, VisionEngine, and LLMsVerifier.

## Key Constraints

1. **Thread safety** -- CoverageTracker and DocGraph use sync.RWMutex. Never bypass locking.
2. **No test deletion** -- Tests must never be removed or disabled. Fix root causes.
3. **Race-safe** -- All tests must pass with `go test ./... -race -count=1`.
4. **Interface stability** -- LLMAgent, Loader, CoverageTracker, FeatureMapBuilder interfaces are public API.
5. **No circular imports** -- pkg/llm has no internal dependencies. pkg/feature depends on pkg/llm and pkg/docgraph.

## Adding New Functionality

1. Write tests first (TDD)
2. Implement the minimum code to pass tests
3. Run `go test ./... -race -count=1`
4. Run `go vet ./...`
5. Ensure SPDX headers on all new files

## Common Tasks

- **Add new document format**: Add parser in `pkg/loader/`, update `DefaultLoader.LoadFile()` switch
- **Add new feature category**: Add constant in `pkg/feature/feature.go`, update `AllCategories()`
- **Add new prompt template**: Add function in `pkg/llm/prompts.go`
- **Add coverage metric**: Add method to `CoverageTracker` interface and `tracker` implementation
