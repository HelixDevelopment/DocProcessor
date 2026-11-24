// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package docprocessor_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"digital.vasic.docprocessor/pkg/coverage"
	"digital.vasic.docprocessor/pkg/feature"
	"digital.vasic.docprocessor/pkg/llm"
	"digital.vasic.docprocessor/pkg/loader"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAgent implements llm.LLMAgent for E2E testing.
type mockAgent struct{}

func (m *mockAgent) Summarize(_ context.Context, text string) (string, error) {
	return "Summary of: " + text[:min(50, len(text))], nil
}

func (m *mockAgent) ExtractFeatures(_ context.Context, _ string) ([]llm.RawFeature, error) {
	return []llm.RawFeature{
		{Name: "Test Feature", Description: "A test feature", Category: "ui", Platforms: []string{"android"}, Priority: "medium"},
	}, nil
}

func (m *mockAgent) ClassifyFeature(_ context.Context, _ llm.RawFeature) (llm.FeatureCategory, error) {
	return "ui", nil
}

func (m *mockAgent) InferScreens(_ context.Context, features []llm.Feature) ([]llm.ExpectedScreen, error) {
	var screens []llm.ExpectedScreen
	for _, f := range features {
		screens = append(screens, llm.ExpectedScreen{
			ID:        "screen-" + f.ID,
			Name:      f.Name + " Screen",
			Platforms: f.Platforms,
			Features:  []string{f.ID},
		})
	}
	return screens, nil
}

func (m *mockAgent) GenerateTestSteps(_ context.Context, f llm.Feature) ([]llm.TestStep, error) {
	return []llm.TestStep{
		{Order: 1, Action: "Open " + f.Name, Expected: "Screen visible", Description: "Open the feature screen"},
		{Order: 2, Action: "Verify display", Expected: "Content correct", Description: "Check content"},
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestE2E_FullPipeline(t *testing.T) {
	// Setup: create a project directory with documentation
	dir := t.TempDir()
	docsDir := filepath.Join(dir, "docs")
	err := os.MkdirAll(docsDir, 0755)
	require.NoError(t, err)

	// Create documentation files
	err = os.WriteFile(filepath.Join(dir, "README.md"), []byte(`# Sample Text Editor

## Overview

Sample Text Editor is a cross-platform text editor supporting 17 text formats with cloud storage integration.

## Features

### Markdown Editing

Full markdown support with syntax highlighting and live preview capabilities.

### CSV Import/Export

Import and export CSV files with delimiter detection and header row support.

### Cloud Storage

Sync documents with Dropbox, Google Drive, and OneDrive cloud storage providers.
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(docsDir, "architecture.md"), []byte(`# Architecture

## Module Structure

The shared module contains all business logic organized by feature type.

## Format System

The format registry provides lazy-loaded parsers for all supported text formats.

See [README](../README.md) for quick start guide.
`), 0644)
	require.NoError(t, err)

	// Phase 1: Load documents
	ctx := context.Background()
	l := loader.NewDefaultLoader([]string{"md", "yaml", "html"})
	docs, err := l.LoadDir(ctx, dir)
	require.NoError(t, err)
	assert.True(t, len(docs) >= 2, "should load at least 2 documents")

	// Phase 2: Build feature map
	builder := feature.NewBuilder(dir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	require.NoError(t, err)
	assert.True(t, len(fm.Features) > 0, "should extract features")
	assert.NotNil(t, fm.DocGraph)

	// Phase 3: Enrich with mock LLM agent
	agent := &mockAgent{}
	err = builder.Enrich(ctx, fm, agent)
	require.NoError(t, err)
	assert.True(t, len(fm.Screens) > 0, "should infer screens")

	// Phase 4: Coverage tracking
	tracker := coverage.NewTracker()
	platforms := []string{"android", "desktop", "web"}

	for _, f := range fm.Features {
		tracker.RegisterFeature(
			coverage.Feature{ID: f.ID, Name: f.Name, Category: string(f.Category)},
			platforms,
		)
	}

	report := tracker.Coverage()
	assert.Equal(t, len(fm.Features)*len(platforms), report.Total)
	assert.Equal(t, 0.0, report.OverallPct)

	// Simulate verification
	for _, f := range fm.Features {
		tracker.MarkVerified(f.ID, "android", coverage.Evidence{
			ScreenshotPath: fmt.Sprintf("/evidence/%s-android.png", f.ID),
			Timestamp:      time.Now(),
		})
	}

	report = tracker.Coverage()
	assert.True(t, report.Verified > 0)
	assert.True(t, tracker.CoverageByPlatform("android") > 0)

	// Verify doc graph has links
	assert.True(t, fm.DocGraph.NodeCount() > 0)

	// Export coverage snapshot
	snapshot := tracker.Export()
	assert.NotEmpty(t, snapshot.Features)
}

func TestE2E_EmptyProject(t *testing.T) {
	dir := t.TempDir()

	ctx := context.Background()
	l := loader.NewDefaultLoader([]string{"md"})
	docs, err := l.LoadDir(ctx, dir)
	require.NoError(t, err)
	assert.Len(t, docs, 0)

	builder := feature.NewBuilder(dir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	require.NoError(t, err)
	assert.Empty(t, fm.Features)
}

func TestE2E_DocGraphExport(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "doc1.md"), []byte(`# Doc 1

See [Doc 2](./doc2.md) and [External](https://example.com).

This has enough content to generate a feature entry.
`), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "doc2.md"), []byte(`# Doc 2

See [Doc 1](./doc1.md).

This also has enough content to generate a feature entry.
`), 0644)
	require.NoError(t, err)

	ctx := context.Background()
	l := loader.NewDefaultLoader([]string{"md"})
	docs, err := l.LoadDir(ctx, dir)
	require.NoError(t, err)

	builder := feature.NewBuilder(dir)
	fm, err := builder.BuildFromDocs(ctx, docs)
	require.NoError(t, err)

	// Export to JSON
	jsonData, err := fm.DocGraph.ExportJSON()
	require.NoError(t, err)
	assert.True(t, len(jsonData) > 0)

	// Export to Mermaid
	mermaid := fm.DocGraph.ExportMermaid()
	assert.Contains(t, mermaid, "graph LR")
	assert.Contains(t, mermaid, "-->")
}

func TestE2E_CoverageByPlatformAndCategory(t *testing.T) {
	tracker := coverage.NewTracker()

	// Register features across categories and platforms
	features := []struct {
		id        string
		category  string
		platforms []string
	}{
		{"feat-md", "format", []string{"android", "desktop", "web"}},
		{"feat-csv", "format", []string{"android", "desktop"}},
		{"feat-theme", "ui", []string{"android", "desktop", "web"}},
		{"feat-cloud", "network", []string{"android", "desktop"}},
	}

	for _, f := range features {
		tracker.RegisterFeature(
			coverage.Feature{ID: f.id, Name: f.id, Category: f.category},
			f.platforms,
		)
	}

	// Verify android features
	tracker.MarkVerified("feat-md", "android", coverage.Evidence{})
	tracker.MarkVerified("feat-csv", "android", coverage.Evidence{})
	tracker.MarkVerified("feat-theme", "android", coverage.Evidence{})
	tracker.MarkVerified("feat-cloud", "android", coverage.Evidence{})

	// Verify desktop features
	tracker.MarkVerified("feat-md", "desktop", coverage.Evidence{})

	assert.Equal(t, 1.0, tracker.CoverageByPlatform("android"))
	assert.InDelta(t, 0.25, tracker.CoverageByPlatform("desktop"), 0.01)
	assert.Equal(t, 0.0, tracker.CoverageByPlatform("web"))

	// Format category: feat-md (3 platforms) + feat-csv (2 platforms) = 5 items, 3 verified
	formatCov := tracker.CoverageByCategory("format")
	assert.InDelta(t, 0.6, formatCov, 0.01)
}

func TestE2E_FeatureMapMerge(t *testing.T) {
	builder := feature.NewBuilder("/project")

	// Build two feature maps from different doc sets
	fm1, err := builder.BuildFromDocs(context.Background(), []loader.Document{
		{
			Path:  "/project/README.md",
			Title: "README",
			Sections: []loader.Section{
				{Title: "Feature A", Content: "Feature A description with enough text to pass heuristic."},
			},
		},
	})
	require.NoError(t, err)

	fm2, err := builder.BuildFromDocs(context.Background(), []loader.Document{
		{
			Path:  "/project/GUIDE.md",
			Title: "Guide",
			Sections: []loader.Section{
				{Title: "Feature B", Content: "Feature B description with enough text to pass heuristic."},
			},
		},
	})
	require.NoError(t, err)

	merged := builder.Merge(fm1, fm2)
	assert.True(t, len(merged.Features) >= 2, "merged map should have features from both sources")
}
