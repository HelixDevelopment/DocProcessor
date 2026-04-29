// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package coverage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConcurrentMarkVerified(t *testing.T) {
	tracker := NewTracker()

	// Register 100 features on 3 platforms
	for i := 0; i < 100; i++ {
		tracker.RegisterFeature(
			Feature{ID: fmt.Sprintf("f%d", i), Name: fmt.Sprintf("Feature %d", i), Category: "ui"},
			[]string{"android", "desktop", "web"},
		)
	}

	var wg sync.WaitGroup
	platforms := []string{"android", "desktop", "web"}

	// Concurrently mark all features as verified on all platforms
	for _, platform := range platforms {
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(featureID, p string) {
				defer wg.Done()
				tracker.MarkVerified(featureID, p, Evidence{
					Timestamp: time.Now(),
				})
			}(fmt.Sprintf("f%d", i), platform)
		}
	}

	wg.Wait()

	report := tracker.Coverage()
	assert.Equal(t, 300, report.Total)
	assert.Equal(t, 300, report.Verified)
	assert.Equal(t, 1.0, report.OverallPct)
}

func TestConcurrentMixedOperations(t *testing.T) {
	tracker := NewTracker()

	for i := 0; i < 30; i++ {
		tracker.RegisterFeature(
			Feature{ID: fmt.Sprintf("f%d", i), Category: "ui"},
			[]string{"android", "desktop"},
		)
	}

	var wg sync.WaitGroup

	// Writers: mark verified
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tracker.MarkVerified(fmt.Sprintf("f%d", id), "android", Evidence{})
		}(i)
	}

	// Writers: mark failed
	for i := 10; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tracker.MarkFailed(fmt.Sprintf("f%d", id), "android", Issue{Title: "Bug"})
		}(i)
	}

	// Writers: mark skipped
	for i := 20; i < 30; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tracker.MarkSkipped(fmt.Sprintf("f%d", id), "android", "N/A")
		}(i)
	}

	// Readers: coverage queries
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tracker.Coverage()
			_ = tracker.CoverageByPlatform("android")
			_ = tracker.CoverageByCategory("ui")
			_ = tracker.Unverified()
		}()
	}

	wg.Wait()

	report := tracker.Coverage()
	assert.Equal(t, 60, report.Total)
}

func TestConcurrentExport(t *testing.T) {
	// bluff-scan: no-assert-ok (concurrency test — go test -race catches data races; absence of panic == correctness)
	tracker := NewTracker()
	for i := 0; i < 20; i++ {
		tracker.RegisterFeature(
			Feature{ID: fmt.Sprintf("f%d", i), Category: "ui"},
			[]string{"android"},
		)
	}

	var wg sync.WaitGroup

	// Concurrent exports while writing
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			tracker.MarkVerified(fmt.Sprintf("f%d", id), "android", Evidence{})
		}(i)
		go func() {
			defer wg.Done()
			_ = tracker.Export()
		}()
	}

	wg.Wait()
}

func TestConcurrentRegisterAndQuery(t *testing.T) {
	tracker := NewTracker()

	var wg sync.WaitGroup

	// Register features concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tracker.RegisterFeature(
				Feature{ID: fmt.Sprintf("f%d", id), Category: "format"},
				[]string{"android"},
			)
		}(i)
	}

	// Query concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tracker.Coverage()
			_ = tracker.Unverified()
		}()
	}

	wg.Wait()

	report := tracker.Coverage()
	assert.Equal(t, 50, report.Total)
}

func TestHighContention(t *testing.T) {
	tracker := NewTracker()
	tracker.RegisterFeature(
		Feature{ID: "f1", Category: "ui"},
		[]string{"android"},
	)

	var wg sync.WaitGroup

	// Many goroutines all operating on the same feature
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if id%3 == 0 {
				tracker.MarkVerified("f1", "android", Evidence{})
			} else if id%3 == 1 {
				tracker.MarkFailed("f1", "android", Issue{})
			} else {
				_ = tracker.Coverage()
			}
		}(i)
	}

	wg.Wait()

	// State should be consistent
	report := tracker.Coverage()
	assert.Equal(t, 1, report.Total)
	// Exactly one of verified, failed, skipped, or unverified
	total := report.Verified + report.Failed + report.Skipped + report.Unverified
	assert.Equal(t, 1, total)
}
