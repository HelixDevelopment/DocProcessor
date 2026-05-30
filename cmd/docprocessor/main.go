// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package main provides the CLI entry point for DocProcessor.
package main

import (
	"context"
	"fmt"
	"os"

	"digital.vasic.docprocessor/pkg/config"
	"digital.vasic.docprocessor/pkg/feature"
	"digital.vasic.docprocessor/pkg/loader"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: docprocessor <docs-directory>\n")
		os.Exit(1)
	}

	docsDir := os.Args[1]
	cfg := config.DefaultConfig()

	l := loader.NewDefaultLoader(cfg.Formats)
	ctx := context.Background()

	docs, err := l.LoadDir(ctx, docsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d documents\n", len(docs))

	builder := feature.NewBuilder(docsDir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building feature map: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Feature map: %d features, %d screens, %d workflows\n",
		len(fm.Features), len(fm.Screens), len(fm.Workflows))
	fmt.Printf("Doc graph: %d nodes, %d edges\n",
		fm.DocGraph.NodeCount(), fm.DocGraph.EdgeCount())

	for cat, features := range fm.Categories {
		fmt.Printf("  Category %s: %d features\n", cat, len(features))
	}
	for platform, features := range fm.PlatformMatrix {
		fmt.Printf("  Platform %s: %d features\n", platform, len(features))
	}
}
