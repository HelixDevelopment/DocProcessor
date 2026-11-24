// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

// Package docgraph provides a directed graph of inter-document links
// with JSON and Mermaid export capabilities.
package docgraph

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Node represents a document in the graph.
type Node struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Edge represents a directed link between two documents.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// DocGraph is a directed graph of document links.
type DocGraph struct {
	nodes map[string]*Node
	edges []Edge
	mu    sync.RWMutex
}

// New creates an empty DocGraph.
func New() *DocGraph {
	return &DocGraph{
		nodes: make(map[string]*Node),
	}
}

// AddNode adds a document node to the graph. If a node with the same ID
// already exists, its title is updated.
func (g *DocGraph) AddNode(id, title string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[id] = &Node{ID: id, Title: title}
}

// AddEdge adds a directed edge from one document to another.
// Both nodes are created if they don't exist (with empty titles).
func (g *DocGraph) AddEdge(from, to string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.nodes[from]; !ok {
		g.nodes[from] = &Node{ID: from, Title: ""}
	}
	if _, ok := g.nodes[to]; !ok {
		g.nodes[to] = &Node{ID: to, Title: ""}
	}

	// Check for duplicate edges
	for _, e := range g.edges {
		if e.From == from && e.To == to {
			return
		}
	}

	g.edges = append(g.edges, Edge{From: from, To: to})
}

// Nodes returns all nodes in the graph.
func (g *DocGraph) Nodes() []Node {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		result = append(result, *n)
	}
	return result
}

// Edges returns all edges in the graph.
func (g *DocGraph) Edges() []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Edge, len(g.edges))
	copy(result, g.edges)
	return result
}

// NodeCount returns the number of nodes.
func (g *DocGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the number of edges.
func (g *DocGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.edges)
}

// HasNode checks if a node exists.
func (g *DocGraph) HasNode(id string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.nodes[id]
	return ok
}

// HasEdge checks if a directed edge exists.
func (g *DocGraph) HasEdge(from, to string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, e := range g.edges {
		if e.From == from && e.To == to {
			return true
		}
	}
	return false
}

// Neighbors returns all nodes linked from the given node.
func (g *DocGraph) Neighbors(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []string
	for _, e := range g.edges {
		if e.From == id {
			result = append(result, e.To)
		}
	}
	return result
}

// IncomingEdges returns all nodes that link to the given node.
func (g *DocGraph) IncomingEdges(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var result []string
	for _, e := range g.edges {
		if e.To == id {
			result = append(result, e.From)
		}
	}
	return result
}

// GraphSnapshot is a serializable snapshot of the graph.
type GraphSnapshot struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// Export returns a serializable snapshot.
func (g *DocGraph) Export() GraphSnapshot {
	return GraphSnapshot{
		Nodes: g.Nodes(),
		Edges: g.Edges(),
	}
}

// ExportJSON returns the graph as a JSON byte slice.
func (g *DocGraph) ExportJSON() ([]byte, error) {
	snapshot := g.Export()
	return json.MarshalIndent(snapshot, "", "  ")
}

// ExportMermaid returns the graph as a Mermaid diagram string.
func (g *DocGraph) ExportMermaid() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("graph LR\n")

	// sanitizeMermaidID is lossy by construction: every non-alphanumeric,
	// non-underscore byte collapses to '_', so two DISTINCT original node
	// IDs that differ only in punctuation (e.g. "docs/api-guide.md" and
	// "docs/api_guide.md") produce the IDENTICAL base Mermaid id. Emitting
	// that colliding id for both would silently conflate two unrelated
	// documents into a single visual node and scramble which edges belong
	// to which document — a content-corrupting bug, not a cosmetic one.
	// mermaidIDs (below) guarantees a unique Mermaid id per distinct
	// original node ID by appending a numeric disambiguator on collision,
	// and the exact same mapping is reused for both node declarations and
	// edges so the exported diagram stays internally consistent. Nodes are
	// also visited in sorted-ID order so the disambiguation outcome (and
	// the rest of the rendered diagram) is deterministic across runs,
	// rather than depending on Go's randomized map iteration order.
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	mermaidIDs := make(map[string]string, len(ids))
	used := make(map[string]bool, len(ids))
	for _, id := range ids {
		base := sanitizeMermaidID(id)
		final := base
		for n := 2; used[final]; n++ {
			final = fmt.Sprintf("%s_%d", base, n)
		}
		used[final] = true
		mermaidIDs[id] = final
	}

	for _, id := range ids {
		n := g.nodes[id]
		label := n.Title
		if label == "" {
			label = n.ID
		}
		// Sanitize label for Mermaid
		label = strings.ReplaceAll(label, "\"", "'")
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", mermaidIDs[id], label))
	}

	for _, e := range g.edges {
		from, ok := mermaidIDs[e.From]
		if !ok {
			// Defensive only: AddEdge always creates both endpoint nodes,
			// so every edge's From/To is present in mermaidIDs above.
			from = sanitizeMermaidID(e.From)
		}
		to, ok := mermaidIDs[e.To]
		if !ok {
			to = sanitizeMermaidID(e.To)
		}
		sb.WriteString(fmt.Sprintf("    %s --> %s\n", from, to))
	}

	return sb.String()
}

// sanitizeMermaidID converts a path/string to a valid Mermaid node ID.
func sanitizeMermaidID(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}
	id := result.String()
	if id == "" {
		return "node"
	}
	return id
}

// ImportJSON loads a graph from a JSON snapshot.
func ImportJSON(data []byte) (*DocGraph, error) {
	var snapshot GraphSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("docgraph: import json: %w", err)
	}
	g := New()
	for _, n := range snapshot.Nodes {
		g.AddNode(n.ID, n.Title)
	}
	for _, e := range snapshot.Edges {
		g.AddEdge(e.From, e.To)
	}
	return g, nil
}
