// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package feature

import (
	"context"
	"fmt"
	"unicode/utf8"

	"digital.vasic.docprocessor/pkg/llm"
	"digital.vasic.docprocessor/pkg/loader"
)

// FeatureMapBuilder defines the interface for building feature maps from documents.
type FeatureMapBuilder interface {
	// BuildFromDocs builds a feature map from loaded documents.
	BuildFromDocs(ctx context.Context, docs []loader.Document) (*FeatureMap, error)

	// Enrich uses an LLM agent to enrich a feature map with inferred screens and test steps.
	Enrich(ctx context.Context, fm *FeatureMap, agent llm.LLMAgent) error

	// Merge combines multiple feature maps into one, deduplicating by feature ID.
	Merge(maps ...*FeatureMap) *FeatureMap
}

// DefaultBuilder implements FeatureMapBuilder.
type DefaultBuilder struct {
	projectRoot string
}

// NewBuilder creates a new DefaultBuilder.
func NewBuilder(projectRoot string) *DefaultBuilder {
	return &DefaultBuilder{projectRoot: projectRoot}
}

// BuildFromDocs scans documents and extracts features from section content.
// Without an LLM agent, it uses heuristic extraction.
func (b *DefaultBuilder) BuildFromDocs(ctx context.Context, docs []loader.Document) (*FeatureMap, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fm := NewFeatureMap(b.projectRoot)

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Add document to doc graph
		fm.DocGraph.AddNode(doc.Path, doc.Title)
		for _, link := range doc.Links {
			fm.DocGraph.AddEdge(doc.Path, link)
		}

		// Extract features from sections
		for _, section := range doc.Sections {
			features := extractFeaturesHeuristic(section, doc.Path)
			for _, f := range features {
				fm.AddFeature(f)
			}
		}
	}

	return fm, nil
}

// Enrich uses an LLM agent to enrich the feature map.
func (b *DefaultBuilder) Enrich(ctx context.Context, fm *FeatureMap, agent llm.LLMAgent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if agent == nil {
		return fmt.Errorf("builder: LLM agent is nil")
	}

	// Infer screens from features
	llmFeatures := toLLMFeatures(fm.Features)
	screens, err := agent.InferScreens(ctx, llmFeatures)
	if err != nil {
		return fmt.Errorf("builder: infer screens: %w", err)
	}
	fm.Screens = fromLLMExpectedScreens(screens)

	// Generate test steps for each feature
	for i := range fm.Features {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		steps, err := agent.GenerateTestSteps(ctx, toLLMFeature(fm.Features[i]))
		if err != nil {
			// Continue on error; partial enrichment is acceptable
			continue
		}
		fm.Features[i].TestSteps = fromLLMTestSteps(steps)
	}

	return nil
}

// Merge combines multiple feature maps, deduplicating by feature ID.
func (b *DefaultBuilder) Merge(maps ...*FeatureMap) *FeatureMap {
	merged := NewFeatureMap(b.projectRoot)
	seen := make(map[string]bool)

	for _, fm := range maps {
		if fm == nil {
			continue
		}
		for _, f := range fm.Features {
			if !seen[f.ID] {
				seen[f.ID] = true
				merged.AddFeature(f)
			}
		}
		for _, s := range fm.Screens {
			merged.Screens = append(merged.Screens, s)
		}
		for _, w := range fm.Workflows {
			merged.Workflows = append(merged.Workflows, w)
		}
		// Merge doc graph
		if fm.DocGraph != nil {
			for _, node := range fm.DocGraph.Nodes() {
				merged.DocGraph.AddNode(node.ID, node.Title)
			}
			for _, edge := range fm.DocGraph.Edges() {
				merged.DocGraph.AddEdge(edge.From, edge.To)
			}
		}
	}

	return merged
}

// extractFeaturesHeuristic extracts features from a section using simple heuristics.
// This is the fallback when no LLM is available.
func extractFeaturesHeuristic(section loader.Section, docPath string) []Feature {
	if section.Content == "" || section.Title == "" {
		return nil
	}

	// Only create features from sections that look like feature descriptions
	// (at least 20 characters of content)
	if len(section.Content) < 20 {
		return nil
	}

	id := GenerateID(section.Title)
	f := Feature{
		ID:            id,
		Name:          section.Title,
		Description:   truncate(section.Content, 500),
		Category:      CategoryOther,
		Platforms:     []string{"android", "desktop", "web"},
		Priority:      "medium",
		SourceDoc:     docPath,
		SourceSection: section.Title,
	}

	return []Feature{f}
}

// truncate shortens s to at most maxLen runes, appending "..." when it cuts.
// It counts and slices by RUNE, never by byte, so multibyte UTF-8 text
// (e.g. Cyrillic) is never split mid-rune into corrupted output.
func truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen]) + "..."
}
