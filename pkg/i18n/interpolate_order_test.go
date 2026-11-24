// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package i18n

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInterpolate_OrderIndependentNoReSubstitution is a reproduce-first
// RED test for an order-dependent data-corruption defect in interpolate:
// it loops `for k, v := range args` (nondeterministic Go map order) and
// applies strings.ReplaceAll per key. When one arg's VALUE happens to
// contain another arg's placeholder literal (e.g. a Cyrillic file path
// "/корен/{{.name}}/файл"), a later pass re-substitutes that literal,
// producing different output depending on iteration order — corrupting
// the rendered message. The bundle contract (active.sr.yaml header)
// states every {{.token}} is preserved EXACTLY; a placeholder that
// originated from a VALUE must never itself be expanded.
//
// Correct behaviour: each {{.key}} present in the ORIGINAL template is
// replaced by its value exactly once; placeholder-looking text that
// appears inside a substituted value is left verbatim.
func TestInterpolate_OrderIndependentNoReSubstitution(t *testing.T) {
	tmpl := "Учитано {{.name}} из {{.path}}"
	args := map[string]any{
		// path's VALUE contains the literal "{{.name}}" — it must NOT be
		// re-expanded into the name value.
		"path": "/корен/{{.name}}/файл",
		"name": "ИмяФайла",
	}
	const want = "Учитано ИмяФайла из /корен/{{.name}}/файл"

	// Run many times: with the buggy map-iteration implementation, output
	// diverges depending on whether "path" or "name" is visited first.
	for i := 0; i < 50; i++ {
		got := interpolate(tmpl, args)
		require.Equalf(t, want, got,
			"interpolate must replace each original-template placeholder exactly once "+
				"and never re-expand placeholder text that came from a value (iteration %d)", i)
	}
}
