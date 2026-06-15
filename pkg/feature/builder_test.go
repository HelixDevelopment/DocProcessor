// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package feature

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"digital.vasic.docprocessor/pkg/llm"
	"digital.vasic.docprocessor/pkg/loader"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMAgent is a mock implementation of llm.LLMAgent for testing.
type mockLLMAgent struct {
	summarizeFunc     func(ctx context.Context, text string) (string, error)
	extractFunc       func(ctx context.Context, text string) ([]llm.RawFeature, error)
	classifyFunc      func(ctx context.Context, feature llm.RawFeature) (llm.FeatureCategory, error)
	inferScreensFunc  func(ctx context.Context, features []llm.Feature) ([]llm.ExpectedScreen, error)
	generateStepsFunc func(ctx context.Context, feature llm.Feature) ([]llm.TestStep, error)
}

func (m *mockLLMAgent) Summarize(ctx context.Context, text string) (string, error) {
	if m.summarizeFunc != nil {
		return m.summarizeFunc(ctx, text)
	}
	return "summary", nil
}

func (m *mockLLMAgent) ExtractFeatures(ctx context.Context, text string) ([]llm.RawFeature, error) {
	if m.extractFunc != nil {
		return m.extractFunc(ctx, text)
	}
	return nil, nil
}

func (m *mockLLMAgent) ClassifyFeature(ctx context.Context, feature llm.RawFeature) (llm.FeatureCategory, error) {
	if m.classifyFunc != nil {
		return m.classifyFunc(ctx, feature)
	}
	return "other", nil
}

func (m *mockLLMAgent) InferScreens(ctx context.Context, features []llm.Feature) ([]llm.ExpectedScreen, error) {
	if m.inferScreensFunc != nil {
		return m.inferScreensFunc(ctx, features)
	}
	return []llm.ExpectedScreen{
		{ID: "screen-1", Name: "Main Screen", Platforms: []string{"android"}},
	}, nil
}

func (m *mockLLMAgent) GenerateTestSteps(ctx context.Context, feature llm.Feature) ([]llm.TestStep, error) {
	if m.generateStepsFunc != nil {
		return m.generateStepsFunc(ctx, feature)
	}
	return []llm.TestStep{
		{Order: 1, Action: "Open app", Expected: "App opens", Description: "Launch app"},
	}, nil
}

func TestBuilder_BuildFromDocs_Empty(t *testing.T) {
	builder := NewBuilder("/project")
	fm, err := builder.BuildFromDocs(context.Background(), nil)
	require.NoError(t, err)
	assert.Empty(t, fm.Features)
}

func TestBuilder_BuildFromDocs_Markdown(t *testing.T) {
	builder := NewBuilder("/project")
	docs := []loader.Document{
		{
			Path:    "/project/docs/features.md",
			Format:  "md",
			Title:   "Features",
			Content: "# Features\n\n## Markdown Editing\n\nEdit markdown files with syntax highlighting and live preview.",
			Sections: []loader.Section{
				{Title: "Features", Level: 1, Content: "Overview content"},
				{Title: "Markdown Editing", Level: 2, Content: "Edit markdown files with syntax highlighting and live preview."},
			},
			Links: []string{"./other.md"},
		},
	}

	fm, err := builder.BuildFromDocs(context.Background(), docs)
	require.NoError(t, err)

	// Should extract features from sections with sufficient content
	assert.True(t, len(fm.Features) > 0, "should extract at least one feature")
	assert.True(t, fm.DocGraph.HasNode("/project/docs/features.md"))
}

func TestBuilder_BuildFromDocs_LinksToDocGraph(t *testing.T) {
	builder := NewBuilder("/project")
	docs := []loader.Document{
		{
			Path:  "/project/README.md",
			Title: "README",
			Links: []string{"./docs/guide.md", "https://example.com"},
			Sections: []loader.Section{
				{Title: "README", Level: 1, Content: "This is the README with enough content to process."},
			},
		},
	}

	fm, err := builder.BuildFromDocs(context.Background(), docs)
	require.NoError(t, err)

	assert.True(t, fm.DocGraph.HasNode("/project/README.md"))
	assert.True(t, fm.DocGraph.HasEdge("/project/README.md", "./docs/guide.md"))
	assert.True(t, fm.DocGraph.HasEdge("/project/README.md", "https://example.com"))
}

func TestBuilder_BuildFromDocs_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	builder := NewBuilder("/project")
	_, err := builder.BuildFromDocs(ctx, []loader.Document{{Path: "test.md"}})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBuilder_Enrich(t *testing.T) {
	builder := NewBuilder("/project")
	fm := NewFeatureMap("/project")
	fm.AddFeature(Feature{
		ID:        "feat-test",
		Name:      "Test Feature",
		Category:  CategoryUI,
		Platforms: []string{"android"},
	})

	agent := &mockLLMAgent{}
	err := builder.Enrich(context.Background(), fm, agent)
	require.NoError(t, err)

	assert.Len(t, fm.Screens, 1)
	assert.Len(t, fm.Features[0].TestSteps, 1)
}

func TestBuilder_Enrich_NilAgent(t *testing.T) {
	builder := NewBuilder("/project")
	fm := NewFeatureMap("/project")

	err := builder.Enrich(context.Background(), fm, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM agent is nil")
}

func TestBuilder_Enrich_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	builder := NewBuilder("/project")
	fm := NewFeatureMap("/project")
	agent := &mockLLMAgent{}

	err := builder.Enrich(ctx, fm, agent)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestBuilder_Enrich_InferScreensError(t *testing.T) {
	builder := NewBuilder("/project")
	fm := NewFeatureMap("/project")
	fm.AddFeature(Feature{ID: "f1", Name: "F1"})

	agent := &mockLLMAgent{
		inferScreensFunc: func(ctx context.Context, features []llm.Feature) ([]llm.ExpectedScreen, error) {
			return nil, fmt.Errorf("LLM error")
		},
	}

	err := builder.Enrich(context.Background(), fm, agent)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "infer screens")
}

func TestBuilder_Enrich_GenerateStepsError(t *testing.T) {
	builder := NewBuilder("/project")
	fm := NewFeatureMap("/project")
	fm.AddFeature(Feature{ID: "f1", Name: "F1"})

	agent := &mockLLMAgent{
		generateStepsFunc: func(ctx context.Context, feature llm.Feature) ([]llm.TestStep, error) {
			return nil, fmt.Errorf("LLM error")
		},
	}

	err := builder.Enrich(context.Background(), fm, agent)
	require.NoError(t, err) // Partial enrichment is acceptable
	assert.Empty(t, fm.Features[0].TestSteps)
}

func TestBuilder_Merge_NilMaps(t *testing.T) {
	builder := NewBuilder("/project")
	merged := builder.Merge(nil, nil)
	assert.Empty(t, merged.Features)
}

func TestBuilder_Merge_Deduplication(t *testing.T) {
	builder := NewBuilder("/project")

	fm1 := NewFeatureMap("/project")
	fm1.AddFeature(Feature{ID: "f1", Name: "Feature 1", Category: CategoryUI, Platforms: []string{"android"}})
	fm1.AddFeature(Feature{ID: "f2", Name: "Feature 2", Category: CategoryFormat, Platforms: []string{"desktop"}})

	fm2 := NewFeatureMap("/project")
	fm2.AddFeature(Feature{ID: "f1", Name: "Feature 1 Duplicate", Category: CategoryUI, Platforms: []string{"android"}})
	fm2.AddFeature(Feature{ID: "f3", Name: "Feature 3", Category: CategoryNetwork, Platforms: []string{"web"}})

	merged := builder.Merge(fm1, fm2)
	assert.Len(t, merged.Features, 3) // f1, f2, f3 (f1 deduplicated)

	// The first occurrence wins
	f1 := merged.FeatureByID("f1")
	require.NotNil(t, f1)
	assert.Equal(t, "Feature 1", f1.Name)
}

func TestBuilder_Merge_DocGraphsMerged(t *testing.T) {
	builder := NewBuilder("/project")

	fm1 := NewFeatureMap("/project")
	fm1.DocGraph.AddNode("doc1", "Doc 1")

	fm2 := NewFeatureMap("/project")
	fm2.DocGraph.AddNode("doc2", "Doc 2")
	fm2.DocGraph.AddEdge("doc2", "doc1")

	merged := builder.Merge(fm1, fm2)
	assert.True(t, merged.DocGraph.HasNode("doc1"))
	assert.True(t, merged.DocGraph.HasNode("doc2"))
	assert.True(t, merged.DocGraph.HasEdge("doc2", "doc1"))
}

func TestExtractFeaturesHeuristic_EmptySection(t *testing.T) {
	section := loader.Section{Title: "Empty", Content: ""}
	features := extractFeaturesHeuristic(section, "doc.md")
	assert.Empty(t, features)
}

func TestExtractFeaturesHeuristic_ShortContent(t *testing.T) {
	section := loader.Section{Title: "Short", Content: "Too short"}
	features := extractFeaturesHeuristic(section, "doc.md")
	assert.Empty(t, features)
}

func TestExtractFeaturesHeuristic_ValidSection(t *testing.T) {
	section := loader.Section{
		Title:   "Markdown Editing",
		Content: "Edit markdown files with syntax highlighting, live preview, and more.",
	}
	features := extractFeaturesHeuristic(section, "/docs/features.md")
	require.Len(t, features, 1)
	assert.Equal(t, "feat-markdown-editing", features[0].ID)
	assert.Equal(t, "/docs/features.md", features[0].SourceDoc)
	assert.Equal(t, "Markdown Editing", features[0].SourceSection)
	assert.Equal(t, CategoryOther, features[0].Category)
	assert.Equal(t, "medium", features[0].Priority)
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "hello", truncate("hello", 10))
	assert.Equal(t, "hell...", truncate("hello world", 4))
	assert.Equal(t, "", truncate("", 10))
}

// TestTruncate_DoesNotCorruptMultibyteRunes is a reproduce-first RED test for
// the byte-vs-rune truncation bug. truncate() slices by BYTE offset (s[:maxLen]);
// for non-ASCII text (Cyrillic, every char = 2 UTF-8 bytes) a byte-offset cut can
// fall in the middle of a multibyte rune, producing invalid UTF-8 in the
// Feature.Description that BuildFromDocs ships to downstream consumers / serializers.
//
// Cyrillic "Подршка за превод" repeated: each Cyrillic letter is 2 bytes, so an
// odd byte-offset cut splits a rune. With maxLen=5 over a Cyrillic string the byte
// slice s[:5] ends mid-rune.
func TestTruncate_DoesNotCorruptMultibyteRunes(t *testing.T) {
	// "Подршка" — 7 Cyrillic letters = 14 bytes. maxLen=5 (bytes) lands mid-rune.
	cyrillic := "Подршка"
	require.False(t, utf8.ValidString(cyrillic[:5]),
		"precondition: a raw byte-slice at offset 5 of this Cyrillic string IS invalid UTF-8")

	out := truncate(cyrillic, 5)

	// The truncated output (minus the "..." ellipsis) must remain valid UTF-8 —
	// truncation must never split a rune and corrupt the text.
	assert.True(t, utf8.ValidString(out),
		"truncate corrupted multibyte text: produced invalid UTF-8 %q", out)
	body := strings.TrimSuffix(out, "...")
	assert.True(t, utf8.ValidString(body),
		"truncate body is invalid UTF-8 (mid-rune cut): %q", body)
}

// TestBuildFromDocs_CyrillicDescriptionStaysValidUTF8 proves the corruption is
// user-visible end-to-end: a long Cyrillic section description (>500 bytes) gets
// truncated by BuildFromDocs and the resulting Feature.Description must be valid
// UTF-8, not a string ending in a broken half-rune.
func TestBuildFromDocs_CyrillicDescriptionStaysValidUTF8(t *testing.T) {
	// Build a Cyrillic description longer than the 500-byte truncate threshold.
	// "Подршка за превод докумената. " is 30 runes / 53 bytes; repeat to exceed 500 bytes.
	unit := "Подршка за превод докумената на ћирилицу. "
	long := strings.Repeat(unit, 20) // well over 500 bytes
	require.Greater(t, len(long), 500)

	doc := loader.Document{
		Path:  "/docs/feature.md",
		Title: "Документ",
		Sections: []loader.Section{
			{Title: "Превод", Level: 2, Content: long, Line: 1},
		},
	}

	b := NewBuilder("/project")
	fm, err := b.BuildFromDocs(context.Background(), []loader.Document{doc})
	require.NoError(t, err)
	require.Len(t, fm.Features, 1)

	desc := fm.Features[0].Description
	assert.True(t, utf8.ValidString(desc),
		"BuildFromDocs shipped a Feature.Description with corrupted (invalid) UTF-8: %q", desc)
}
