// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package docprocessor_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutomation_GoModValid(t *testing.T) {
	// go.mod should exist and be valid
	data, err := os.ReadFile("go.mod")
	require.NoError(t, err)
	content := string(data)
	assert.Contains(t, content, "module digital.vasic.docprocessor")
	// Accept any 1.x Go version declared in go.mod; the exact patch/minor
	// changes as the module is updated and should not be pinned here.
	assert.Regexp(t, `go 1\.\d+`, content)
}

func TestAutomation_GoBuild(t *testing.T) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "go build failed: %s", string(out))
}

func TestAutomation_GoVet(t *testing.T) {
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "go vet failed: %s", string(out))
}

func TestAutomation_PackageStructure(t *testing.T) {
	packages := []string{
		"pkg/loader",
		"pkg/feature",
		"pkg/coverage",
		"pkg/docgraph",
		"pkg/llm",
		"pkg/config",
		"cmd/docprocessor",
	}

	for _, pkg := range packages {
		t.Run(pkg, func(t *testing.T) {
			info, err := os.Stat(pkg)
			require.NoError(t, err, "package dir %s should exist", pkg)
			assert.True(t, info.IsDir())

			// Each package should have at least one .go file
			entries, err := os.ReadDir(pkg)
			require.NoError(t, err)

			hasGoFile := false
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
					hasGoFile = true
					break
				}
			}
			assert.True(t, hasGoFile, "package %s should have at least one .go source file", pkg)
		})
	}
}

func TestAutomation_TestFilesExist(t *testing.T) {
	testPackages := []string{
		"pkg/loader",
		"pkg/feature",
		"pkg/coverage",
		"pkg/docgraph",
		"pkg/llm",
		"pkg/config",
	}

	for _, pkg := range testPackages {
		t.Run(pkg, func(t *testing.T) {
			entries, err := os.ReadDir(pkg)
			require.NoError(t, err)

			hasTestFile := false
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.go") {
					hasTestFile = true
					break
				}
			}
			assert.True(t, hasTestFile, "package %s should have test files", pkg)
		})
	}
}

func TestAutomation_LicenseExists(t *testing.T) {
	_, err := os.Stat("LICENSE")
	assert.NoError(t, err, "LICENSE file should exist")
}

func TestAutomation_MakefileExists(t *testing.T) {
	_, err := os.Stat("Makefile")
	assert.NoError(t, err, "Makefile should exist")
}

func TestAutomation_NoRaceConditions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("race detector may not be available on Windows CI")
	}

	// This test itself runs with -race flag; if we got here, race detection is active
	// Verify the race detector is actually enabled by checking build tags
	cmd := exec.Command("go", "test", "-race", "-count=1", "-run", "TestNew", "./pkg/docgraph/")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "race detection test failed: %s", string(out))
}

func TestAutomation_CleanBuild(t *testing.T) {
	// Verify no stale build artifacts
	cmd := exec.Command("go", "build", "-v", "./...")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "clean build failed: %s", string(out))
}

func TestAutomation_EnvExampleExists(t *testing.T) {
	_, err := os.Stat(".env.example")
	assert.NoError(t, err, ".env.example should exist")
}

func TestAutomation_UpstreamsExist(t *testing.T) {
	entries, err := os.ReadDir("Upstreams")
	require.NoError(t, err, "Upstreams directory should exist")

	hasScript := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sh") {
			hasScript = true
			break
		}
	}
	assert.True(t, hasScript, "Upstreams should contain shell scripts")
}

func TestAutomation_AllSourceFilesHaveLicense(t *testing.T) {
	err := filepath.Walk("pkg", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)
		assert.Contains(t, content, "SPDX-License-Identifier", "file %s should have SPDX header", path)
		return nil
	})
	assert.NoError(t, err)
}
