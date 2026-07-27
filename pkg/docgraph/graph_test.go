// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package docgraph

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	g := New()
	assert.NotNil(t, g)
	assert.Equal(t, 0, g.NodeCount())
	assert.Equal(t, 0, g.EdgeCount())
}

func TestAddNode(t *testing.T) {
	g := New()
	g.AddNode("doc1", "Document 1")
	assert.Equal(t, 1, g.NodeCount())
	assert.True(t, g.HasNode("doc1"))
}

func TestAddNode_UpdateTitle(t *testing.T) {
	g := New()
	g.AddNode("doc1", "Old Title")
	g.AddNode("doc1", "New Title")
	assert.Equal(t, 1, g.NodeCount())

	nodes := g.Nodes()
	found := false
	for _, n := range nodes {
		if n.ID == "doc1" {
			assert.Equal(t, "New Title", n.Title)
			found = true
		}
	}
	assert.True(t, found)
}

func TestAddEdge(t *testing.T) {
	g := New()
	g.AddNode("doc1", "Doc 1")
	g.AddNode("doc2", "Doc 2")
	g.AddEdge("doc1", "doc2")

	assert.Equal(t, 1, g.EdgeCount())
	assert.True(t, g.HasEdge("doc1", "doc2"))
	assert.False(t, g.HasEdge("doc2", "doc1")) // directed
}

func TestAddEdge_AutoCreateNodes(t *testing.T) {
	g := New()
	g.AddEdge("new1", "new2")

	assert.True(t, g.HasNode("new1"))
	assert.True(t, g.HasNode("new2"))
	assert.Equal(t, 2, g.NodeCount())
}

func TestAddEdge_NoDuplicates(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("a", "b")
	g.AddEdge("a", "b")

	assert.Equal(t, 1, g.EdgeCount())
}

func TestNeighbors(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "c")

	neighbors := g.Neighbors("a")
	assert.Len(t, neighbors, 2)
	assert.Contains(t, neighbors, "b")
	assert.Contains(t, neighbors, "c")

	neighbors = g.Neighbors("b")
	assert.Len(t, neighbors, 1)
	assert.Contains(t, neighbors, "c")

	neighbors = g.Neighbors("c")
	assert.Len(t, neighbors, 0)
}

func TestIncomingEdges(t *testing.T) {
	g := New()
	g.AddEdge("a", "c")
	g.AddEdge("b", "c")

	incoming := g.IncomingEdges("c")
	assert.Len(t, incoming, 2)
	assert.Contains(t, incoming, "a")
	assert.Contains(t, incoming, "b")

	incoming = g.IncomingEdges("a")
	assert.Len(t, incoming, 0)
}

func TestNodes(t *testing.T) {
	g := New()
	g.AddNode("a", "A")
	g.AddNode("b", "B")
	g.AddNode("c", "C")

	nodes := g.Nodes()
	assert.Len(t, nodes, 3)
}

func TestEdges(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("b", "c")

	edges := g.Edges()
	assert.Len(t, edges, 2)
}

func TestExport(t *testing.T) {
	g := New()
	g.AddNode("doc1", "Doc 1")
	g.AddNode("doc2", "Doc 2")
	g.AddEdge("doc1", "doc2")

	snapshot := g.Export()
	assert.Len(t, snapshot.Nodes, 2)
	assert.Len(t, snapshot.Edges, 1)
}

func TestExportJSON(t *testing.T) {
	g := New()
	g.AddNode("doc1", "Doc 1")
	g.AddEdge("doc1", "doc2")

	data, err := g.ExportJSON()
	require.NoError(t, err)

	var snapshot GraphSnapshot
	err = json.Unmarshal(data, &snapshot)
	require.NoError(t, err)
	assert.Len(t, snapshot.Nodes, 2)
	assert.Len(t, snapshot.Edges, 1)
}

func TestExportMermaid(t *testing.T) {
	g := New()
	g.AddNode("readme", "README")
	g.AddNode("guide", "User Guide")
	g.AddEdge("readme", "guide")

	mermaid := g.ExportMermaid()
	assert.Contains(t, mermaid, "graph LR")
	assert.Contains(t, mermaid, "readme")
	assert.Contains(t, mermaid, "guide")
	assert.Contains(t, mermaid, "-->")
}

func TestExportMermaid_SpecialCharacters(t *testing.T) {
	g := New()
	g.AddNode("/path/to/doc.md", "My Doc's Title")

	mermaid := g.ExportMermaid()
	assert.Contains(t, mermaid, "graph LR")
	// Quotes in titles should be escaped
	assert.NotContains(t, mermaid, `"My Doc"s Title"`)
}

func TestImportJSON(t *testing.T) {
	original := New()
	original.AddNode("a", "Node A")
	original.AddNode("b", "Node B")
	original.AddEdge("a", "b")

	data, err := original.ExportJSON()
	require.NoError(t, err)

	imported, err := ImportJSON(data)
	require.NoError(t, err)

	assert.Equal(t, original.NodeCount(), imported.NodeCount())
	assert.Equal(t, original.EdgeCount(), imported.EdgeCount())
	assert.True(t, imported.HasNode("a"))
	assert.True(t, imported.HasNode("b"))
	assert.True(t, imported.HasEdge("a", "b"))
}

func TestImportJSON_Invalid(t *testing.T) {
	_, err := ImportJSON([]byte("invalid json"))
	assert.Error(t, err)
}

func TestImportJSON_Empty(t *testing.T) {
	g, err := ImportJSON([]byte(`{"nodes":[],"edges":[]}`))
	require.NoError(t, err)
	assert.Equal(t, 0, g.NodeCount())
	assert.Equal(t, 0, g.EdgeCount())
}

func TestHasNode_Nonexistent(t *testing.T) {
	g := New()
	assert.False(t, g.HasNode("nonexistent"))
}

func TestHasEdge_Nonexistent(t *testing.T) {
	g := New()
	assert.False(t, g.HasEdge("a", "b"))
}

func TestSanitizeMermaidID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"/path/to/file.md", "_path_to_file_md"},
		{"hello world", "hello_world"},
		{"", "node"},
		{"abc123", "abc123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeMermaidID(tt.input))
		})
	}
}

func TestExportMermaid_EmptyGraph(t *testing.T) {
	g := New()
	mermaid := g.ExportMermaid()
	assert.Equal(t, "graph LR\n", mermaid)
}

// TestExportMermaid_DistinctIDsMustNotCollide is a reproduce-first RED test
// for the Mermaid-ID-collision class: sanitizeMermaidID maps every
// non-alphanumeric byte to '_', so two DISTINCT original node IDs whose
// only difference is punctuation (a hyphen vs an underscore, a slash vs a
// dot) sanitize to the IDENTICAL Mermaid node id. "docs/api-guide.md" and
// "docs/api_guide.md" are two real, independently-added documents in this
// graph, yet both sanitize to "docs_api_guide_md".
//
// Before the fix, ExportMermaid emitted TWO node-declaration lines sharing
// that one colliding id — Mermaid silently renders only one visual node for
// what are two distinct documents, and worse, an edge meant for the SECOND
// document renders as if it belongs to the FIRST (or vice versa) because
// both endpoints resolve to the same id. This is silent content corruption
// of the generated diagram, not a cosmetic quirk — exactly the transform
// gap the anti-bluff covenant forbids (a "transform" that silently merges
// distinct content).
func TestExportMermaid_DistinctIDsMustNotCollide(t *testing.T) {
	g := New()
	g.AddNode("docs/api-guide.md", "API Guide (hyphen)")
	g.AddNode("docs/api_guide.md", "API Guide (underscore)")
	g.AddEdge("docs/api-guide.md", "docs/other.md")
	g.AddEdge("docs/api_guide.md", "docs/other.md")

	// Precondition: the two original IDs really do collide under the raw
	// sanitizer — if this ever stops being true the test is testing the
	// wrong scenario.
	require.Equal(t, sanitizeMermaidID("docs/api-guide.md"), sanitizeMermaidID("docs/api_guide.md"),
		"precondition: these two distinct doc IDs must collide under the raw per-string sanitizer")

	mermaid := g.ExportMermaid()

	// The exported diagram MUST declare one node line per distinct graph
	// node (3 nodes: the two api-guide docs + "docs/other.md"). Count the
	// node-declaration lines (the `id["label"]` lines, not edge lines).
	nodeLines := 0
	for _, line := range strings.Split(mermaid, "\n") {
		if strings.Contains(line, "[\"") {
			nodeLines++
		}
	}
	assert.Equal(t, g.NodeCount(), nodeLines,
		"ExportMermaid must emit exactly one node-declaration line per distinct node; "+
			"a lower count means two distinct nodes were silently merged under a colliding Mermaid id:\n%s", mermaid)

	// Both original node labels must be individually present and distinctly
	// addressable — not merged into a single visual node.
	assert.Contains(t, mermaid, "API Guide (hyphen)")
	assert.Contains(t, mermaid, "API Guide (underscore)")

	// Both edges (api-guide-hyphen -> other, api-guide-underscore -> other)
	// must render as two DISTINCT edge lines with two DISTINCT source ids —
	// not two edges that both accidentally read as coming from the same
	// (colliding) source id.
	var edgeLines []string
	for _, line := range strings.Split(mermaid, "\n") {
		if strings.Contains(line, "-->") {
			edgeLines = append(edgeLines, strings.TrimSpace(line))
		}
	}
	require.Len(t, edgeLines, 2, "expected exactly 2 edge lines:\n%s", mermaid)
	assert.NotEqual(t, edgeLines[0], edgeLines[1],
		"the two edges must be distinguishable by source id, not collapsed to the same line:\n%s", mermaid)
}

func TestNeighbors_NonexistentNode(t *testing.T) {
	g := New()
	neighbors := g.Neighbors("nonexistent")
	assert.Empty(t, neighbors)
}

func TestIncomingEdges_NonexistentNode(t *testing.T) {
	g := New()
	incoming := g.IncomingEdges("nonexistent")
	assert.Empty(t, incoming)
}

func TestExportJSON_RoundTrip_LargeGraph(t *testing.T) {
	g := New()
	for i := 0; i < 50; i++ {
		id := strings.Repeat("n", 1) + string(rune('A'+i%26))
		g.AddNode(id, "Node "+id)
	}
	// Add some edges
	nodes := g.Nodes()
	for i := 0; i < len(nodes)-1; i++ {
		g.AddEdge(nodes[i].ID, nodes[i+1].ID)
	}

	data, err := g.ExportJSON()
	require.NoError(t, err)

	imported, err := ImportJSON(data)
	require.NoError(t, err)

	assert.Equal(t, g.NodeCount(), imported.NodeCount())
	assert.Equal(t, g.EdgeCount(), imported.EdgeCount())
}
