// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package main provides the CLI entry point for DocProcessor.
package main

import (
	"context"
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

// main wires the production Translator and delegates to runCLI, the
// real, testable CLI body. main() itself carries no behaviour beyond
// dependency wiring + process-exit translation so every code path that
// matters is exercised by main_test.go through runCLI.
func main() {
	tr := newTranslator()
	os.Exit(runCLI(context.Background(), os.Args, os.Stdout, os.Stderr, tr))
}

// newTranslator returns the production Translator.
//
// The CLI message IDs runCLI emits use the `docprocessor_cli_*`
// namespace and live in the FLAT-schema bundle pkg/i18n/bundles/
// active.<locale>.yaml. As of the pkg/i18n flat-bundle support
// (NewBundleTranslator now loads the `active.*` surface in a second
// pass — see pkg/i18n/bundle.go), the BundleTranslator resolves every
// `docprocessor_cli_*` ID to its real localised string, so the
// production CLI renders genuine text rather than id-echo. The bundles
// are embedded via `//go:embed bundles/*.yaml`, so resolution is
// CWD-independent in the installed binary.
//
// If bundle construction fails for any reason, fall back to
// NoopTranslator (id-echo) so the binary stays functional and never
// panics — the honest degraded mode. This is NEVER a silent
// empty-string substitution (which would be a §11.4 PASS-bluff at the
// i18n layer); the id-echo fallback keeps every line visible.
func newTranslator() i18n.Translator {
	tr, err := i18n.NewBundleTranslator("en")
	if err != nil {
		return i18n.NoopTranslator{}
	}
	return tr
}

// runCLI is the real CLI body. It returns the process exit code so the
// caller (main) and tests can both observe the exact outcome. All
// user-facing strings are resolved through the supplied Translator per
// CONST-046; no hardcoded English literal is printed.
func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer, tr i18n.Translator) int {
	start := time.Now()

	verbose, positional := parseArgs(args[1:])

	if len(positional) < 1 {
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_usage", nil))
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_help_header", nil))
		return 1
	}

	docsDir := positional[0]
	if strings.TrimSpace(docsDir) == "" {
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_path_invalid", map[string]any{"path": docsDir}))
		return 1
	}

	absDir, err := filepath.Abs(docsDir)
	if err != nil {
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_resolving_path", map[string]any{"path": docsDir, "error": err.Error()}))
		return 1
	}

	cfg := config.DefaultConfig()
	fprintln(stdout, tr.T(ctx, "docprocessor_cli_format_summary", map[string]any{"formats": strings.Join(cfg.Formats, ", ")}))

	l := loader.NewDefaultLoader(cfg.Formats)
	docs, err := l.LoadDir(ctx, absDir)
	if err != nil {
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_loading_docs", map[string]any{"error": err.Error()}))
		return 1
	}

	fprintln(stdout, tr.T(ctx, "docprocessor_cli_loaded_documents", map[string]any{"count": len(docs)}))

	if len(docs) == 0 {
		fprintln(stdout, tr.T(ctx, "docprocessor_cli_no_docs_found", map[string]any{"path": absDir}))
		fprintln(stdout, tr.T(ctx, "docprocessor_cli_done", map[string]any{"elapsed_ms": time.Since(start).Milliseconds()}))
		return 0
	}

	builder := feature.NewBuilder(absDir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	if err != nil {
		fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_building_feature_map", map[string]any{"error": err.Error()}))
		return 1
	}

	fprintln(stdout, tr.T(ctx, "docprocessor_cli_summary_header", nil))
	fprintln(stdout, tr.T(ctx, "docprocessor_cli_feature_map_summary", map[string]any{
		"features":  len(fm.Features),
		"screens":   len(fm.Screens),
		"workflows": len(fm.Workflows),
	}))
	fprintln(stdout, tr.T(ctx, "docprocessor_cli_doc_graph_summary", map[string]any{
		"nodes": fm.DocGraph.NodeCount(),
		"edges": fm.DocGraph.EdgeCount(),
	}))

	for cat, features := range fm.Categories {
		fprintln(stdout, tr.T(ctx, "docprocessor_cli_category_line", map[string]any{
			"category": string(cat),
			"count":    len(features),
		}))
	}
	for platform, features := range fm.PlatformMatrix {
		fprintln(stdout, tr.T(ctx, "docprocessor_cli_platform_line", map[string]any{
			"platform": platform,
			"count":    len(features),
		}))
	}

	if verbose {
		for _, f := range fm.Features {
			fprintln(stdout, tr.T(ctx, "docprocessor_cli_feature_line", map[string]any{
				"id":       f.ID,
				"name":     f.Name,
				"category": string(f.Category),
			}))
		}
		for _, s := range fm.Screens {
			fprintln(stdout, tr.T(ctx, "docprocessor_cli_screen_line", map[string]any{
				"id":   s.ID,
				"name": s.Name,
			}))
		}
		for _, w := range fm.Workflows {
			fprintln(stdout, tr.T(ctx, "docprocessor_cli_workflow_line", map[string]any{
				"id":    w.ID,
				"name":  w.Name,
				"steps": len(w.Steps),
			}))
		}
	}

	fprintln(stdout, tr.T(ctx, "docprocessor_cli_done", map[string]any{"elapsed_ms": time.Since(start).Milliseconds()}))
	return 0
}

// parseArgs splits the post-program-name argument slice into the
// --verbose flag and the remaining positional arguments. Unknown
// flags are treated as positionals so an explicit docs-directory
// starting with a dash is not silently dropped.
func parseArgs(args []string) (verbose bool, positional []string) {
	for _, a := range args {
		if a == "--verbose" || a == "-v" {
			verbose = true
			continue
		}
		positional = append(positional, a)
	}
	return verbose, positional
}

// fprintln writes s followed by a newline, ignoring the write error
// (stdout/stderr write failures are not actionable from a CLI body).
func fprintln(w io.Writer, s string) {
	_, _ = io.WriteString(w, s+"\n")
}
