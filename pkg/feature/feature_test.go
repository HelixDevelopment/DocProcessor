// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package feature

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateID_Basic(t *testing.T) {
	id := GenerateID("Markdown Editing")
	assert.Equal(t, "feat-markdown-editing", id)
}

func TestGenerateID_SpecialCharacters(t *testing.T) {
	id := GenerateID("CSV/TSV Import & Export")
	assert.Equal(t, "feat-csvtsv-import--export", id)
}

func TestGenerateID_Deterministic(t *testing.T) {
	id1 := GenerateID("Feature X")
	id2 := GenerateID("Feature X")
	assert.Equal(t, id1, id2)
}

func TestGenerateID_DifferentNames(t *testing.T) {
	id1 := GenerateID("Feature A")
	id2 := GenerateID("Feature B")
	assert.NotEqual(t, id1, id2)
}

func TestGenerateID_LongName(t *testing.T) {
	longName := "This is a very long feature name that exceeds the maximum slug length and should be truncated"
	id := GenerateID(longName)
	assert.True(t, len(id) < 60, "ID should be truncated for long names")
	assert.True(t, len(id) > 5, "ID should not be empty")
}

func TestGenerateID_EmptyName(t *testing.T) {
	id := GenerateID("")
	assert.Equal(t, "feat-", id)
}

func TestGenerateID_WhitespaceOnly(t *testing.T) {
	id := GenerateID("   ")
	assert.Equal(t, "feat-", id)
}

func TestGenerateID_CaseInsensitive(t *testing.T) {
	id1 := GenerateID("Feature X")
	id2 := GenerateID("feature x")
	assert.Equal(t, id1, id2)
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	assert.Len(t, cats, 8)
	assert.Contains(t, cats, CategoryFormat)
	assert.Contains(t, cats, CategoryUI)
	assert.Contains(t, cats, CategoryNetwork)
	assert.Contains(t, cats, CategorySettings)
	assert.Contains(t, cats, CategoryStorage)
	assert.Contains(t, cats, CategoryAuth)
	assert.Contains(t, cats, CategoryEditor)
	assert.Contains(t, cats, CategoryOther)
}

func TestValidCategory(t *testing.T) {
	assert.True(t, ValidCategory("format"))
	assert.True(t, ValidCategory("ui"))
	assert.True(t, ValidCategory("other"))
	assert.False(t, ValidCategory("invalid"))
	assert.False(t, ValidCategory(""))
}

func TestNewFeatureMap(t *testing.T) {
	fm := NewFeatureMap("/project/root")
	assert.Equal(t, "/project/root", fm.ProjectRoot)
	assert.NotNil(t, fm.Categories)
	assert.NotNil(t, fm.PlatformMatrix)
	assert.NotNil(t, fm.DocGraph)
	assert.Empty(t, fm.Features)
}

func TestFeatureMap_AddFeature(t *testing.T) {
	fm := NewFeatureMap("/root")

	f := Feature{
		ID:        "feat-test",
		Name:      "Test Feature",
		Category:  CategoryUI,
		Platforms: []string{"android", "desktop"},
	}
	fm.AddFeature(f)

	assert.Len(t, fm.Features, 1)
	assert.Len(t, fm.Categories[CategoryUI], 1)
	assert.Len(t, fm.PlatformMatrix["android"], 1)
	assert.Len(t, fm.PlatformMatrix["desktop"], 1)
}

func TestFeatureMap_AddMultipleFeatures(t *testing.T) {
	fm := NewFeatureMap("/root")

	fm.AddFeature(Feature{ID: "f1", Name: "F1", Category: CategoryUI, Platforms: []string{"android"}})
	fm.AddFeature(Feature{ID: "f2", Name: "F2", Category: CategoryUI, Platforms: []string{"android", "desktop"}})
	fm.AddFeature(Feature{ID: "f3", Name: "F3", Category: CategoryFormat, Platforms: []string{"desktop"}})

	assert.Len(t, fm.Features, 3)
	assert.Len(t, fm.Categories[CategoryUI], 2)
	assert.Len(t, fm.Categories[CategoryFormat], 1)
	assert.Len(t, fm.PlatformMatrix["android"], 2)
	assert.Len(t, fm.PlatformMatrix["desktop"], 2)
}

func TestFeatureMap_FeatureByID(t *testing.T) {
	fm := NewFeatureMap("/root")
	fm.AddFeature(Feature{ID: "feat-test", Name: "Test"})

	f := fm.FeatureByID("feat-test")
	require.NotNil(t, f)
	assert.Equal(t, "Test", f.Name)

	f2 := fm.FeatureByID("nonexistent")
	assert.Nil(t, f2)
}

func TestFeatureMap_FeaturesForPlatform(t *testing.T) {
	fm := NewFeatureMap("/root")
	fm.AddFeature(Feature{ID: "f1", Platforms: []string{"android"}})
	fm.AddFeature(Feature{ID: "f2", Platforms: []string{"android", "web"}})
	fm.AddFeature(Feature{ID: "f3", Platforms: []string{"desktop"}})

	android := fm.FeaturesForPlatform("android")
	assert.Len(t, android, 2)

	web := fm.FeaturesForPlatform("web")
	assert.Len(t, web, 1)

	ios := fm.FeaturesForPlatform("ios")
	assert.Len(t, ios, 0)
}

func TestFeatureMap_FeaturesForCategory(t *testing.T) {
	fm := NewFeatureMap("/root")
	fm.AddFeature(Feature{ID: "f1", Category: CategoryUI})
	fm.AddFeature(Feature{ID: "f2", Category: CategoryUI})
	fm.AddFeature(Feature{ID: "f3", Category: CategoryFormat})

	ui := fm.FeaturesForCategory(CategoryUI)
	assert.Len(t, ui, 2)

	format := fm.FeaturesForCategory(CategoryFormat)
	assert.Len(t, format, 1)

	network := fm.FeaturesForCategory(CategoryNetwork)
	assert.Len(t, network, 0)
}

func TestFeature_Struct(t *testing.T) {
	f := Feature{
		ID:            "feat-markdown",
		Name:          "Markdown Editing",
		Description:   "Edit markdown files",
		Category:      CategoryFormat,
		Platforms:     []string{"android", "desktop", "web"},
		Priority:      "high",
		Screens:       []string{"screen-editor"},
		SourceDoc:     "/docs/features.md",
		SourceSection: "Markdown",
		TestSteps: []TestStep{
			{Order: 1, Action: "Open file", Expected: "Editor opens"},
		},
	}

	assert.Equal(t, "feat-markdown", f.ID)
	assert.Equal(t, CategoryFormat, f.Category)
	assert.Len(t, f.Platforms, 3)
	assert.Len(t, f.TestSteps, 1)
}

func TestExpectedScreen_Struct(t *testing.T) {
	s := ExpectedScreen{
		ID:          "screen-editor",
		Name:        "Editor Screen",
		Description: "Main text editing screen",
		Features:    []string{"feat-markdown", "feat-csv"},
		Platforms:   []string{"android", "desktop"},
	}

	assert.Equal(t, "screen-editor", s.ID)
	assert.Len(t, s.Features, 2)
}

func TestWorkflow_Struct(t *testing.T) {
	w := Workflow{
		ID:          "wf-open-edit-save",
		Name:        "Open, Edit, Save",
		Description: "Basic editing workflow",
		Steps: []WorkflowStep{
			{Order: 1, Description: "Open file", ScreenID: "screen-picker"},
			{Order: 2, Description: "Edit text", ScreenID: "screen-editor"},
			{Order: 3, Description: "Save file", ScreenID: "screen-editor"},
		},
		Features:  []string{"feat-open", "feat-edit", "feat-save"},
		Platforms: []string{"android", "desktop", "web"},
	}

	assert.Equal(t, "wf-open-edit-save", w.ID)
	assert.Len(t, w.Steps, 3)
	assert.Len(t, w.Features, 3)
}
