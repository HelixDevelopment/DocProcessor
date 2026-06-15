// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package loader

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultLoader(t *testing.T) {
	l := NewDefaultLoader([]string{"md", "yaml"})
	assert.Equal(t, []string{"md", "yaml"}, l.SupportedFormats())
}

func TestLoadFile_Markdown(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "test.md")
	content := `# My Document

## Overview

This is an overview section with enough content to be meaningful.

## Features

- Feature 1
- Feature 2

See [link1](https://example.com) and [link2](./other.md).
`
	err := os.WriteFile(mdFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), mdFile)
	require.NoError(t, err)

	assert.Equal(t, "My Document", doc.Title)
	assert.Equal(t, "md", doc.Format)
	assert.Equal(t, content, doc.Content)
	assert.Len(t, doc.Sections, 3) // h1 + 2 h2
	assert.Equal(t, "My Document", doc.Sections[0].Title)
	assert.Equal(t, 1, doc.Sections[0].Level)
	assert.Equal(t, "Overview", doc.Sections[1].Title)
	assert.Equal(t, 2, doc.Sections[1].Level)
	assert.Equal(t, "Features", doc.Sections[2].Title)

	assert.Contains(t, doc.Links, "https://example.com")
	assert.Contains(t, doc.Links, "./other.md")
	assert.Len(t, doc.Links, 2)
}

func TestLoadFile_YAML(t *testing.T) {
	dir := t.TempDir()
	yamlFile := filepath.Join(dir, "config.yaml")
	content := `title: Test Config
version: 1.0
features:
  - markdown
  - yaml
`
	err := os.WriteFile(yamlFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"yaml"})
	doc, err := l.LoadFile(context.Background(), yamlFile)
	require.NoError(t, err)

	assert.Equal(t, "Test Config", doc.Title)
	assert.Equal(t, "yaml", doc.Format)
	assert.True(t, len(doc.Sections) > 0)
}

func TestLoadFile_YML_Extension(t *testing.T) {
	dir := t.TempDir()
	ymlFile := filepath.Join(dir, "config.yml")
	content := `name: Test
`
	err := os.WriteFile(ymlFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"yaml"})
	doc, err := l.LoadFile(context.Background(), ymlFile)
	require.NoError(t, err)

	assert.Equal(t, "Test", doc.Title)
	assert.Equal(t, "yaml", doc.Format)
}

func TestLoadFile_HTML(t *testing.T) {
	dir := t.TempDir()
	htmlFile := filepath.Join(dir, "page.html")
	content := `<html>
<head><title>Test</title></head>
<body>Hello</body>
</html>`
	err := os.WriteFile(htmlFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"html"})
	doc, err := l.LoadFile(context.Background(), htmlFile)
	require.NoError(t, err)

	assert.Equal(t, "html", doc.Format)
	assert.Equal(t, content, doc.Content)
}

func TestLoadFile_UnsupportedFormat(t *testing.T) {
	dir := t.TempDir()
	txtFile := filepath.Join(dir, "test.txt")
	err := os.WriteFile(txtFile, []byte("hello"), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	_, err = l.LoadFile(context.Background(), txtFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestLoadFile_FileNotFound(t *testing.T) {
	l := NewDefaultLoader([]string{"md"})
	_, err := l.LoadFile(context.Background(), "/nonexistent/file.md")
	assert.Error(t, err)
}

func TestLoadFile_Directory(t *testing.T) {
	dir := t.TempDir()
	l := NewDefaultLoader([]string{"md"})
	_, err := l.LoadFile(context.Background(), dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestLoadFile_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	l := NewDefaultLoader([]string{"md"})
	_, err := l.LoadFile(ctx, "/any/path.md")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestLoadDir_Basic(t *testing.T) {
	dir := t.TempDir()

	// Create some test files
	err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# README\n\nContent here is long enough to pass."), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("title: Config\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("not loaded"), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md", "yaml"})
	docs, err := l.LoadDir(context.Background(), dir)
	require.NoError(t, err)

	assert.Len(t, docs, 2)
}

func TestLoadDir_Recursive(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "sub")
	err := os.MkdirAll(subdir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root\n\nSome content here."), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subdir, "nested.md"), []byte("# Nested\n\nNested content here."), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	docs, err := l.LoadDir(context.Background(), dir)
	require.NoError(t, err)

	assert.Len(t, docs, 2)
}

func TestLoadDir_SkipsHiddenDirectories(t *testing.T) {
	dir := t.TempDir()
	hiddenDir := filepath.Join(dir, ".hidden")
	err := os.MkdirAll(hiddenDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "visible.md"), []byte("# Visible\n\nSome content here."), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(hiddenDir, "hidden.md"), []byte("# Hidden\n\nHidden content."), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	docs, err := l.LoadDir(context.Background(), dir)
	require.NoError(t, err)

	assert.Len(t, docs, 1)
	assert.Equal(t, "Visible", docs[0].Title)
}

func TestLoadDir_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	l := NewDefaultLoader([]string{"md"})
	docs, err := l.LoadDir(context.Background(), dir)
	require.NoError(t, err)

	assert.Len(t, docs, 0)
}

func TestLoadDir_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.md")
	err := os.WriteFile(file, []byte("# Test"), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	_, err = l.LoadDir(context.Background(), file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestLoadDir_NonexistentPath(t *testing.T) {
	l := NewDefaultLoader([]string{"md"})
	_, err := l.LoadDir(context.Background(), "/nonexistent/path")
	assert.Error(t, err)
}

func TestLoadDir_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	l := NewDefaultLoader([]string{"md"})
	_, err := l.LoadDir(ctx, "/any/path")
	assert.ErrorIs(t, err, context.Canceled)
}

func TestParseMarkdown_NoHeadings(t *testing.T) {
	content := "Just some text without any headings."
	title, sections, links := parseMarkdown(content)
	assert.Empty(t, title)
	assert.Len(t, sections, 0)
	assert.Len(t, links, 0)
}

func TestParseMarkdown_MultipleH1(t *testing.T) {
	content := "# First Title\n\nContent.\n\n# Second Title\n\nMore content."
	title, sections, _ := parseMarkdown(content)
	assert.Equal(t, "First Title", title) // First h1
	assert.Len(t, sections, 2)
}

func TestParseMarkdown_NestedHeadings(t *testing.T) {
	content := "# H1\n\n## H2\n\n### H3\n\n#### H4\n\n##### H5\n\n###### H6"
	_, sections, _ := parseMarkdown(content)
	assert.Len(t, sections, 6)
	for i, s := range sections {
		assert.Equal(t, i+1, s.Level)
	}
}

// TestParseMarkdown_FencedCodeHeadingNotASection is a §11.4.115 RED-baseline-on-the-
// broken-artifact regression guard for the fenced-code-block heading defect:
// a `#`-prefixed line INSIDE a ``` fenced code block (e.g. a shell comment) was
// wrongly parsed as a Markdown heading/section, which (a) fabricated a phantom
// section titled with the code-comment text, (b) truncated the real preceding
// section's Content at the opening fence (genuine content loss), and (c)
// misattributed the prose after the code block to the phantom section.
func TestParseMarkdown_FencedCodeHeadingNotASection(t *testing.T) {
	content := "# Real Title\n\nIntro text.\n\n```sh\n# this is a shell comment, not a heading\necho hi\n```\n\nMore intro after code.\n"
	title, sections, _ := parseMarkdown(content)

	assert.Equal(t, "Real Title", title)
	// Exactly ONE real heading exists; the `#` inside the fence must NOT become a section.
	require.Len(t, sections, 1, "fenced-code `#` line must not be parsed as a heading")
	assert.Equal(t, "Real Title", sections[0].Title)
	// The phantom comment-title must never appear.
	for _, s := range sections {
		assert.NotContains(t, s.Title, "shell comment", "code-comment text leaked into a section title")
	}
	// No content loss: the real section's Content must retain the prose that
	// followed the code block AND the code block itself.
	assert.Contains(t, sections[0].Content, "More intro after code.", "prose after code block was lost/misattributed")
	assert.Contains(t, sections[0].Content, "echo hi", "code-block body was lost from section content")
}

// TestParseMarkdown_TildeFencedCodeHeadingNotASection covers the `~~~` fence variant.
func TestParseMarkdown_TildeFencedCodeHeadingNotASection(t *testing.T) {
	content := "# Title\n\n~~~\n## not a heading inside tilde fence\n~~~\n\nafter.\n"
	_, sections, _ := parseMarkdown(content)
	require.Len(t, sections, 1)
	assert.Equal(t, "Title", sections[0].Title)
	assert.Contains(t, sections[0].Content, "after.")
}

func TestParseMarkdown_DeduplicateLinks(t *testing.T) {
	content := "[link](https://example.com) and [link](https://example.com) again"
	_, _, links := parseMarkdown(content)
	assert.Len(t, links, 1)
}

func TestParseMarkdown_LineNumbers(t *testing.T) {
	content := "# Title\n\n## Section\n\nContent.\n\n## Another\n"
	_, sections, _ := parseMarkdown(content)
	assert.Equal(t, 1, sections[0].Line)
	assert.Equal(t, 3, sections[1].Line)
	assert.Equal(t, 7, sections[2].Line)
}

func TestParseYAML_Valid(t *testing.T) {
	content := `title: My Config
version: 1.0`
	title, sections, _, err := parseYAML(content)
	assert.NoError(t, err)
	assert.Equal(t, "My Config", title)
	assert.True(t, len(sections) > 0)
}

func TestParseYAML_NameFallback(t *testing.T) {
	content := `name: Named Config
version: 2.0`
	title, _, _, err := parseYAML(content)
	assert.NoError(t, err)
	assert.Equal(t, "Named Config", title)
}

func TestParseYAML_Invalid(t *testing.T) {
	content := `[invalid yaml`
	_, _, _, err := parseYAML(content)
	assert.Error(t, err)
}

func TestParseYAML_EmptyDoc(t *testing.T) {
	content := ``
	title, sections, links, err := parseYAML(content)
	// Empty YAML is valid but produces no data
	if err != nil {
		// Some YAML parsers error on empty, which is acceptable
		return
	}
	assert.Empty(t, title)
	assert.Empty(t, sections)
	assert.Empty(t, links)
}

func TestSupportedFormats(t *testing.T) {
	formats := []string{"md", "yaml", "html", "adoc", "rst"}
	l := NewDefaultLoader(formats)
	assert.Equal(t, formats, l.SupportedFormats())
}

func TestLoadFile_EmptyMarkdown(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "empty.md")
	err := os.WriteFile(mdFile, []byte(""), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), mdFile)
	require.NoError(t, err)

	// Title falls back to filename
	assert.Equal(t, "empty", doc.Title)
	assert.Empty(t, doc.Content)
}

func TestLoadFile_AdocFormat(t *testing.T) {
	dir := t.TempDir()
	adocFile := filepath.Join(dir, "doc.adoc")
	content := "= AsciiDoc Title\n\nContent here."
	err := os.WriteFile(adocFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"adoc"})
	doc, err := l.LoadFile(context.Background(), adocFile)
	require.NoError(t, err)

	assert.Equal(t, "adoc", doc.Format)
	assert.Equal(t, "= AsciiDoc Title", doc.Title)
}

func TestLoadFile_RstFormat(t *testing.T) {
	dir := t.TempDir()
	rstFile := filepath.Join(dir, "doc.rst")
	content := "RST Title\n=========\n\nContent here."
	err := os.WriteFile(rstFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"rst"})
	doc, err := l.LoadFile(context.Background(), rstFile)
	require.NoError(t, err)

	assert.Equal(t, "rst", doc.Format)
}
