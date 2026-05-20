// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package i18n

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed bundles/*.yaml
var bundleFS embed.FS

// message is one localised entry. Either Other (singular / generic)
// is set, or the plural forms One / Other are used for plural
// selection.
type message struct {
	One   string `yaml:"one"`
	Other string `yaml:"other"`
}

// BundleTranslator is the default Translator: it loads per-locale
// YAML message bundles embedded in this package and interpolates
// {{.key}} placeholders at render time. It satisfies CONST-046 by
// keeping every user-facing CLI string in a locale-aware resource
// file rather than a Go literal.
//
// Locale selection: the requested locale falls back to its base
// language ("sr-RS" -> "sr"), then to the configured default locale.
// A still-unresolved messageID is echoed verbatim (loud, never
// silent) per the package contract.
type BundleTranslator struct {
	mu            sync.RWMutex
	defaultLocale string
	locales       map[string]map[string]message
}

// localeKey is the context key under which a caller may store the
// active locale string. Using an unexported type avoids collisions.
type localeKey struct{}

// WithLocale returns a context carrying the active locale (e.g.
// "sr", "ja", "en"). DocProcessor reads it in T / TPlural so a
// single Translator instance serves every operator language.
func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeKey{}, locale)
}

// localeFromContext extracts the locale stored by WithLocale, or ""
// when none was set.
func localeFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(localeKey{}).(string); ok {
		return v
	}
	return ""
}

// NewBundleTranslator loads every embedded YAML bundle and returns a
// ready Translator. defaultLocale is used when a request carries no
// locale or an unknown one. An error is returned when the embedded
// bundles cannot be parsed (a build-integrity failure, never a
// runtime user condition).
func NewBundleTranslator(defaultLocale string) (*BundleTranslator, error) {
	bt := &BundleTranslator{
		defaultLocale: defaultLocale,
		locales:       make(map[string]map[string]message),
	}
	entries, err := bundleFS.ReadDir("bundles")
	if err != nil {
		return nil, fmt.Errorf("i18n: read bundles dir: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		raw, err := bundleFS.ReadFile("bundles/" + name)
		if err != nil {
			return nil, fmt.Errorf("i18n: read bundle %s: %w", name, err)
		}
		var msgs map[string]message
		if err := yaml.Unmarshal(raw, &msgs); err != nil {
			return nil, fmt.Errorf("i18n: parse bundle %s: %w", name, err)
		}
		locale := strings.TrimSuffix(name, ".yaml")
		bt.locales[locale] = msgs
	}
	if _, ok := bt.locales[defaultLocale]; !ok {
		return nil, fmt.Errorf("i18n: default locale %q has no bundle", defaultLocale)
	}
	return bt, nil
}

// lookup resolves a message for the given locale, applying base-language
// then default-locale fallback. The bool reports whether a bundle entry
// was found at all.
func (bt *BundleTranslator) lookup(locale, id string) (message, bool) {
	bt.mu.RLock()
	defer bt.mu.RUnlock()

	candidates := []string{locale}
	if i := strings.IndexAny(locale, "-_"); i > 0 {
		candidates = append(candidates, locale[:i])
	}
	candidates = append(candidates, bt.defaultLocale)

	for _, c := range candidates {
		if c == "" {
			continue
		}
		if msgs, ok := bt.locales[c]; ok {
			if m, ok := msgs[id]; ok {
				return m, true
			}
		}
	}
	return message{}, false
}

// interpolate substitutes {{.key}} placeholders using args.
func interpolate(tmpl string, args map[string]any) string {
	if len(args) == 0 || !strings.Contains(tmpl, "{{") {
		return tmpl
	}
	out := tmpl
	for k, v := range args {
		out = strings.ReplaceAll(out, "{{."+k+"}}", fmt.Sprint(v))
	}
	return out
}

// T resolves messageID for the locale carried by ctx (or the default
// locale). An unknown messageID is echoed verbatim.
func (bt *BundleTranslator) T(ctx context.Context, messageID string, args map[string]any) string {
	m, ok := bt.lookup(localeFromContext(ctx), messageID)
	if !ok {
		return messageID
	}
	return interpolate(m.Other, args)
}

// TPlural resolves messageID with English-style plural selection
// (count == 1 -> one, else other). args carries non-count
// placeholders; the count itself is exposed as {{.count}}.
func (bt *BundleTranslator) TPlural(ctx context.Context, messageID string, count int, args map[string]any) string {
	m, ok := bt.lookup(localeFromContext(ctx), messageID)
	if !ok {
		return messageID
	}
	form := m.Other
	if count == 1 && m.One != "" {
		form = m.One
	}
	merged := make(map[string]any, len(args)+1)
	for k, v := range args {
		merged[k] = v
	}
	merged["count"] = count
	return interpolate(form, merged)
}
