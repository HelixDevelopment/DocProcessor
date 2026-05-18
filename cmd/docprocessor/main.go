// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package main provides the CLI entry point for DocProcessor.
//
// CONST-046 (no-hardcoded-content) compliance: every user-facing line
// emitted to stdout/stderr is resolved through a pkg/i18n.Translator
// (see ./pkg/i18n/translator.go) so consuming projects may wire their
// own locale-aware implementation without touching DocProcessor.
//
// runCLI is exported (lower-case but called from main + tests in
// package main) so the wire-level output is testable in-process via
// io.Writer injection — round 97 unit tests assert sentinel
// translation strings appear, not the historical English literals.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"digital.vasic.docprocessor/pkg/config"
	"digital.vasic.docprocessor/pkg/feature"
	"digital.vasic.docprocessor/pkg/i18n"
	"digital.vasic.docprocessor/pkg/loader"
)

func main() {
	tr := i18n.NoopTranslator{}
	exitCode := runCLI(context.Background(), os.Args, os.Stdout, os.Stderr, tr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

// runCLI is the testable body of the docprocessor CLI. It returns the
// shell exit code (0 on success, 1 on failure) instead of calling
// os.Exit directly so tests can drive it with synthetic args and a
// fake Translator while asserting against captured stdout/stderr.
func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer, tr i18n.Translator) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_usage", nil))
		return 1
	}

	docsDir := args[1]
	cfg := config.DefaultConfig()

	l := loader.NewDefaultLoader(cfg.Formats)

	docs, err := l.LoadDir(ctx, docsDir)
	if err != nil {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_loading_docs", map[string]any{
			"error": err.Error(),
		}))
		return 1
	}

	fmt.Fprintln(stdout, tr.T(ctx, "docprocessor_cli_loaded_documents", map[string]any{
		"count": len(docs),
	}))

	builder := feature.NewBuilder(docsDir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	if err != nil {
		fmt.Fprintln(stderr, tr.T(ctx, "docprocessor_cli_error_building_feature_map", map[string]any{
			"error": err.Error(),
		}))
		return 1
	}

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

	return 0
}
