// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package main provides the CLI entry point for DocProcessor.
//
<<<<<<< HEAD
// CONST-046 (no-hardcoded-content) compliance: every user-facing line
// emitted to stdout/stderr is resolved through a pkg/i18n.Translator
// (see ./pkg/i18n/translator.go) so consuming projects may wire their
// own locale-aware implementation without touching DocProcessor.
//
// runCLI is exported (lower-case but called from main + tests in
// package main) so the wire-level output is testable in-process via
// io.Writer injection — round 97 unit tests assert sentinel
// translation strings appear, not the historical English literals.
//
// Round 209 (2026-05-19) — second migration wave. The original round
// 97 batch externalised 8 hardcoded CLI lines (usage, error-loading,
// loaded-count, error-building, feature-map-summary, doc-graph-summary,
// category-line, platform-line). Round 209 extends the CLI surface
// with 10 ADDITIONAL user-facing message IDs covering UX gaps the
// original output did not address:
//
//   1.  docprocessor_cli_help_header          — startup banner
//   2.  docprocessor_cli_path_invalid         — empty/invalid arg path
//   3.  docprocessor_cli_error_resolving_path — distinct path-resolution failure
//   4.  docprocessor_cli_no_docs_found        — 0-document case (was silent)
//   5.  docprocessor_cli_format_summary       — supported-formats line
//   6.  docprocessor_cli_summary_header       — section heading for summary
//   7.  docprocessor_cli_feature_line         — per-feature line (verbose mode)
//   8.  docprocessor_cli_screen_line          — per-screen line (verbose mode)
//   9.  docprocessor_cli_workflow_line        — per-workflow line (verbose mode)
//   10. docprocessor_cli_done                 — completion line with elapsed ms
//
// All ten go through the Translator indirection per CONST-046 — there
// are NO hardcoded English literals in the wire output. Tests assert
// sentinel strings + forbidden-literal absence per round 97 pattern.
//
// Verbatim 2026-05-19 operator mandate (per CONST-049 §11.4.17): "all
// existing tests and Challenges do work in anti-bluff manner - they
// MUST confirm that all tested codebase really works as expected!"
=======
// Every user-facing line this binary prints is sourced from the
// locale-aware i18n bundle (pkg/i18n) per CONST-046 — no English
// string literal is embedded in the print calls. Operators select
// their language via the DOCPROCESSOR_LOCALE environment variable
// (e.g. DOCPROCESSOR_LOCALE=sr); when unset, English ("en") is used.
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"digital.vasic.docprocessor/pkg/config"
	"digital.vasic.docprocessor/pkg/feature"
	"digital.vasic.docprocessor/pkg/i18n"
	"digital.vasic.docprocessor/pkg/loader"
)

// localeEnvVar names the environment variable an operator sets to
// pick the CLI output language. Keeping locale selection in an env
// var (config injection) keeps DocProcessor project-not-aware per
// CONST-051(B).
const localeEnvVar = "DOCPROCESSOR_LOCALE"

// defaultLocale is the bundle used when DOCPROCESSOR_LOCALE is unset
// or names a locale with no bundle.
const defaultLocale = "en"

func main() {
<<<<<<< HEAD
	tr := i18n.NoopTranslator{}
	exitCode := runCLI(context.Background(), os.Args, os.Stdout, os.Stderr, tr)
	if exitCode != 0 {
		os.Exit(exitCode)
=======
	tr, err := i18n.NewBundleTranslator(defaultLocale)
	if err != nil {
		// A bundle-load failure is a build-integrity defect, not a
		// user condition; surface it raw and abort.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.Background()
	if locale := os.Getenv(localeEnvVar); locale != "" {
		ctx = i18n.WithLocale(ctx, locale)
	}

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, tr.T(ctx, "cli_usage", nil))
		os.Exit(1)
	}

	docsDir := os.Args[1]
	cfg := config.DefaultConfig()

	l := loader.NewDefaultLoader(cfg.Formats)

	docs, err := l.LoadDir(ctx, docsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, tr.T(ctx, "cli_error_loading_docs", map[string]any{"err": err}))
		os.Exit(1)
	}

	fmt.Println(tr.TPlural(ctx, "cli_docs_loaded", len(docs), nil))

	builder := feature.NewBuilder(docsDir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	if err != nil {
		fmt.Fprintln(os.Stderr, tr.T(ctx, "cli_error_building_feature_map", map[string]any{"err": err}))
		os.Exit(1)
	}

	fmt.Println(tr.T(ctx, "cli_feature_map_summary", map[string]any{
		"features":  len(fm.Features),
		"screens":   len(fm.Screens),
		"workflows": len(fm.Workflows),
	}))
	fmt.Println(tr.T(ctx, "cli_doc_graph_summary", map[string]any{
		"nodes": fm.DocGraph.NodeCount(),
		"edges": fm.DocGraph.EdgeCount(),
	}))

	for cat, features := range fm.Categories {
		fmt.Println(tr.TPlural(ctx, "cli_category_line", len(features), map[string]any{
			"category": cat,
		}))
	}
	for platform, features := range fm.PlatformMatrix {
		fmt.Println(tr.TPlural(ctx, "cli_platform_line", len(features), map[string]any{
			"platform": platform,
		}))
>>>>>>> 9f5637d2d695cd5fcf8349d1f1b8bf780fa5d865
	}
}

// runCLI is the testable body of the docprocessor CLI. It returns the
// shell exit code (0 on success, 1 on failure) instead of calling
// os.Exit directly so tests can drive it with synthetic args and a
// fake Translator while asserting against captured stdout/stderr.
//
// Args grammar (round 209):
//
//	docprocessor [--verbose] <docs-directory>
//
// --verbose enables per-feature / per-screen / per-workflow lines.
// Without the flag, the summary block stays terse (loaded-count +
// feature-map-summary + doc-graph-summary + category/platform lines
// from round 97).
func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer, tr i18n.Translator) int {
	start := time.Now()

	verbose := false
	positional := make([]string, 0, len(args))
	for i, a := range args {
		if i == 0 {
			positional = append(positional, a)
			continue
		}
		switch a {
		case "--verbose", "-v":
			verbose = true
		default:
			positional = append(positional, a)
		}
	}

	if len(positional) < 2 {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_usage", nil))
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_help_header", nil))
		return 1
	}

	docsDir := positional[1]
	if strings.TrimSpace(docsDir) == "" {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_path_invalid", map[string]any{
			"path": docsDir,
		}))
		return 1
	}

	absDir, err := filepath.Abs(docsDir)
	if err != nil {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_resolving_path", map[string]any{
			"path":  docsDir,
			"error": err.Error(),
		}))
		return 1
	}

	cfg := config.DefaultConfig()

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_format_summary", map[string]any{
		"formats": strings.Join(cfg.Formats, ", "),
	}))

	l := loader.NewDefaultLoader(cfg.Formats)

	docs, err := l.LoadDir(ctx, absDir)
	if err != nil {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_loading_docs", map[string]any{
			"error": err.Error(),
		}))
		return 1
	}

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_loaded_documents", map[string]any{
		"count": len(docs),
	}))

	if len(docs) == 0 {
		fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_no_docs_found", map[string]any{
			"path": absDir,
		}))
		fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_done", map[string]any{
			"elapsed_ms": time.Since(start).Milliseconds(),
		}))
		return 0
	}

	builder := feature.NewBuilder(absDir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	if err != nil {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_building_feature_map", map[string]any{
			"error": err.Error(),
		}))
		return 1
	}

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_summary_header", nil))

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_feature_map_summary", map[string]any{
		"features":  len(fm.Features),
		"screens":   len(fm.Screens),
		"workflows": len(fm.Workflows),
	}))
	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_doc_graph_summary", map[string]any{
		"nodes": fm.DocGraph.NodeCount(),
		"edges": fm.DocGraph.EdgeCount(),
	}))

	for cat, features := range fm.Categories {
		fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_category_line", map[string]any{
			"category": cat,
			"count":    len(features),
		}))
	}
	for platform, features := range fm.PlatformMatrix {
		fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_platform_line", map[string]any{
			"platform": platform,
			"count":    len(features),
		}))
	}

	if verbose {
		for _, f := range fm.Features {
			fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_feature_line", map[string]any{
				"id":       f.ID,
				"name":     f.Name,
				"category": f.Category,
			}))
		}
		for _, s := range fm.Screens {
			fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_screen_line", map[string]any{
				"id":   s.ID,
				"name": s.Name,
			}))
		}
		for _, w := range fm.Workflows {
			fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_workflow_line", map[string]any{
				"id":    w.ID,
				"name":  w.Name,
				"steps": len(w.Steps),
			}))
		}
	}

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_done", map[string]any{
		"elapsed_ms": time.Since(start).Milliseconds(),
	}))

	return 0
}
