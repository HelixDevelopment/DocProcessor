// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package docgraph

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentAddNodes(t *testing.T) {
	g := New()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			g.AddNode(fmt.Sprintf("node-%d", id), fmt.Sprintf("Title %d", id))
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 100, g.NodeCount())
}

func TestConcurrentAddEdges(t *testing.T) {
	g := New()
	// Pre-create nodes
	for i := 0; i < 50; i++ {
		g.AddNode(fmt.Sprintf("node-%d", i), fmt.Sprintf("Title %d", i))
	}

	var wg sync.WaitGroup
	for i := 0; i < 49; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			g.AddEdge(fmt.Sprintf("node-%d", id), fmt.Sprintf("node-%d", id+1))
		}(i)
	}

	wg.Wait()
	assert.Equal(t, 49, g.EdgeCount())
}

func TestConcurrentReadWrite(t *testing.T) {
	// bluff-scan: no-assert-ok (concurrency test — go test -race catches data races; absence of panic == correctness)
	g := New()
	for i := 0; i < 20; i++ {
		g.AddNode(fmt.Sprintf("node-%d", i), fmt.Sprintf("Title %d", i))
	}

	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			g.AddEdge(fmt.Sprintf("node-%d", id), fmt.Sprintf("node-%d", id+10))
		}(i)
	}

	// Readers
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = g.Nodes()
			_ = g.Edges()
			_ = g.NodeCount()
			_ = g.EdgeCount()
		}()
	}

	wg.Wait()
}

func TestConcurrentExport(t *testing.T) {
	// bluff-scan: no-assert-ok (concurrency test — go test -race catches data races; absence of panic == correctness)
	g := New()
	for i := 0; i < 30; i++ {
		g.AddNode(fmt.Sprintf("node-%d", i), fmt.Sprintf("Title %d", i))
		if i > 0 {
			g.AddEdge(fmt.Sprintf("node-%d", i-1), fmt.Sprintf("node-%d", i))
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = g.Export()
			_, _ = g.ExportJSON()
			_ = g.ExportMermaid()
		}()
	}

	wg.Wait()
}

func TestConcurrentNeighbors(t *testing.T) {
	g := New()
	for i := 0; i < 20; i++ {
		g.AddNode(fmt.Sprintf("node-%d", i), fmt.Sprintf("Title %d", i))
		if i > 0 {
			g.AddEdge(fmt.Sprintf("node-%d", 0), fmt.Sprintf("node-%d", i))
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			neighbors := g.Neighbors("node-0")
			assert.Len(t, neighbors, 19)
		}()
	}

	wg.Wait()
}

func TestLargeGraph(t *testing.T) {
	g := New()

	// Create a large graph
	nodeCount := 1000
	for i := 0; i < nodeCount; i++ {
		g.AddNode(fmt.Sprintf("n%d", i), fmt.Sprintf("Title %d", i))
	}
	for i := 0; i < nodeCount-1; i++ {
		g.AddEdge(fmt.Sprintf("n%d", i), fmt.Sprintf("n%d", i+1))
	}

	assert.Equal(t, nodeCount, g.NodeCount())
	assert.Equal(t, nodeCount-1, g.EdgeCount())

	// Export should work
	snapshot := g.Export()
	assert.Len(t, snapshot.Nodes, nodeCount)
	assert.Len(t, snapshot.Edges, nodeCount-1)
}
