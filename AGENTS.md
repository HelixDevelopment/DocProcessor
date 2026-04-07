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

