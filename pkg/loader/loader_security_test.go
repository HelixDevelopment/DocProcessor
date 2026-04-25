// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package loader

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFile_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	// Create a file in the temp dir
	mdFile := filepath.Join(dir, "test.md")
	err := os.WriteFile(mdFile, []byte("# Test\n\nContent."), 0644)
	require.NoError(t, err)

	// Try to access with path traversal (should resolve to absolute)
	l := NewDefaultLoader([]string{"md"})
	_, err = l.LoadFile(context.Background(), filepath.Join(dir, "..", filepath.Base(dir), "test.md"))
	// Should succeed because path gets resolved to the same absolute path
	assert.NoError(t, err)
}

func TestLoadFile_LargeFile(t *testing.T) {
	dir := t.TempDir()
	largeFile := filepath.Join(dir, "large.md")

	// Create a file just over MaxFileSize
	data := strings.Repeat("x", MaxFileSize+1)
	err := os.WriteFile(largeFile, []byte(data), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	_, err = l.LoadFile(context.Background(), largeFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds max size")
}

func TestLoadFile_ExactlyMaxSize(t *testing.T) {
	dir := t.TempDir()
	maxFile := filepath.Join(dir, "max.md")

	data := "# Title\n\n" + strings.Repeat("x", MaxFileSize-10)
	err := os.WriteFile(maxFile, []byte(data), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), maxFile)
	require.NoError(t, err)
	assert.Equal(t, "Title", doc.Title)
}

func TestLoadFile_SymlinkOutsideDir(t *testing.T) {
	dir := t.TempDir()
	outsideDir := t.TempDir()

	// Create a file in the outside directory
	outsideFile := filepath.Join(outsideDir, "secret.md")
	err := os.WriteFile(outsideFile, []byte("# Secret\n\nSecret content."), 0644)
	require.NoError(t, err)

	// Create symlink inside dir pointing outside
	symlink := filepath.Join(dir, "link.md")
	err = os.Symlink(outsideFile, symlink)
	if err != nil {
		t.Skip("symlinks not supported on this platform")  // SKIP-OK: #legacy-untriaged
	}

	// Loader follows symlinks (by design — it resolves to absolute paths)
	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), symlink)
	require.NoError(t, err)
	assert.Equal(t, "Secret", doc.Title)
}

func TestLoadFile_NullBytesInContent(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "null.md")
	content := "# Title\n\nContent with \x00 null bytes.\n"
	err := os.WriteFile(mdFile, []byte(content), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), mdFile)
	require.NoError(t, err)
	assert.Contains(t, doc.Content, "\x00")
}

func TestLoadDir_SymlinkLoop(t *testing.T) {
	dir := t.TempDir()

	// Try to create a symlink loop
	err := os.Symlink(dir, filepath.Join(dir, "loop"))
	if err != nil {
		t.Skip("symlinks not supported on this platform")  // SKIP-OK: #legacy-untriaged
	}

	err = os.WriteFile(filepath.Join(dir, "test.md"), []byte("# Test\n\nContent."), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	// filepath.Walk handles symlink loops by not following symlinks to directories
	docs, err := l.LoadDir(context.Background(), dir)
	// Should not hang; may or may not return error depending on OS
	if err == nil {
		assert.True(t, len(docs) >= 1)
	}
}

func TestLoadFile_SpecialCharactersInFilename(t *testing.T) {
	dir := t.TempDir()
	mdFile := filepath.Join(dir, "my file (1).md")
	err := os.WriteFile(mdFile, []byte("# Test\n\nContent here."), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), mdFile)
	require.NoError(t, err)
	assert.Equal(t, "Test", doc.Title)
}

func TestLoadFile_DeepNestedPath(t *testing.T) {
	dir := t.TempDir()
	deepDir := filepath.Join(dir, "a", "b", "c", "d", "e", "f")
	err := os.MkdirAll(deepDir, 0755)
	require.NoError(t, err)

	mdFile := filepath.Join(deepDir, "deep.md")
	err = os.WriteFile(mdFile, []byte("# Deep\n\nContent here."), 0644)
	require.NoError(t, err)

	l := NewDefaultLoader([]string{"md"})
	doc, err := l.LoadFile(context.Background(), mdFile)
	require.NoError(t, err)
	assert.Equal(t, "Deep", doc.Title)
}
