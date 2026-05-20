// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package main provides the CLI entry point for DocProcessor.
//
// Every user-facing line this binary prints is sourced from the
// locale-aware i18n bundle (pkg/i18n) per CONST-046 — no English
// string literal is embedded in the print calls. Operators select
// their language via the DOCPROCESSOR_LOCALE environment variable
// (e.g. DOCPROCESSOR_LOCALE=sr); when unset, English ("en") is used.
package main

import (
	"context"
	"fmt"
	"os"

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
	}
}
