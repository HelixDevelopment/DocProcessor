// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Milos Vasic

package loader

import (
	"regexp"
	"strings"
)

var (
	markdownHeadingRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	markdownLinkRe    = regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)
)

// parseMarkdown extracts sections and links from Markdown content.
func parseMarkdown(content string) (title string, sections []Section, links []string) {
	// Extract headings as sections. Headings that fall inside a fenced code
	// block (``` or ~~~ ... fence) are NOT real headings — e.g. a shell
	// comment line "# install" inside a ```sh block must stay part of the
	// surrounding section's content, never become a phantom section. Parsing
	// such a line as a heading both fabricates a bogus section and truncates
	// the real section's content at the opening fence (data loss).
	matches := markdownHeadingRe.FindAllStringSubmatchIndex(content, -1)
	matches = filterHeadingsOutsideCodeFences(content, matches)
	for i, loc := range matches {
		level := loc[3] - loc[2] // length of # prefix
		heading := content[loc[4]:loc[5]]

		// Determine section content: from end of heading line to next heading or EOF
		sectionStart := loc[1]
		var sectionEnd int
		if i+1 < len(matches) {
			sectionEnd = matches[i+1][0]
		} else {
			sectionEnd = len(content)
		}
		sectionContent := strings.TrimSpace(content[sectionStart:sectionEnd])

		// Line number (1-based)
		line := strings.Count(content[:loc[0]], "\n") + 1

		sections = append(sections, Section{
			Title:   heading,
			Level:   level,
			Content: sectionContent,
			Line:    line,
		})

		// First h1 is the document title
		if title == "" && level == 1 {
			title = heading
		}
	}

	// Extract links
	linkMatches := markdownLinkRe.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	for _, m := range linkMatches {
		link := m[2]
		if !seen[link] {
			links = append(links, link)
			seen[link] = true
		}
	}

	return title, sections, links
}

// fenceRange is a half-open [start, end) byte interval of content that lies
// inside a fenced code block (the fence lines themselves included).
type fenceRange struct {
	start int
	end   int
}

// codeFenceRanges scans content line-by-line and returns the byte ranges that
// are inside fenced code blocks. A fence opens on a line whose trimmed text
// begins with at least three backticks or three tildes, and closes on the next
// line opening with the same fence character (CommonMark: an unclosed fence
// runs to end of document).
func codeFenceRanges(content string) []fenceRange {
	var ranges []fenceRange
	var (
		inFence   bool
		fenceCh   byte
		blockFrom int
	)
	offset := 0
	for _, line := range strings.SplitAfter(content, "\n") {
		if line == "" {
			break
		}
		trimmed := strings.TrimLeft(line, " \t")
		isBacktick := strings.HasPrefix(trimmed, "```")
		isTilde := strings.HasPrefix(trimmed, "~~~")
		if !inFence {
			if isBacktick || isTilde {
				inFence = true
				if isBacktick {
					fenceCh = '`'
				} else {
					fenceCh = '~'
				}
				blockFrom = offset
			}
		} else {
			closes := (fenceCh == '`' && isBacktick) || (fenceCh == '~' && isTilde)
			if closes {
				inFence = false
				ranges = append(ranges, fenceRange{start: blockFrom, end: offset + len(line)})
			}
		}
		offset += len(line)
	}
	if inFence {
		// Unclosed fence extends to end of document.
		ranges = append(ranges, fenceRange{start: blockFrom, end: len(content)})
	}
	return ranges
}

// filterHeadingsOutsideCodeFences drops heading matches whose start offset lies
// inside any fenced code block, so a `#`/`##`-prefixed line inside a code block
// is never treated as a section heading.
func filterHeadingsOutsideCodeFences(content string, matches [][]int) [][]int {
	fences := codeFenceRanges(content)
	if len(fences) == 0 {
		return matches
	}
	kept := matches[:0:0]
	for _, loc := range matches {
		start := loc[0]
		inside := false
		for _, fr := range fences {
			if start >= fr.start && start < fr.end {
				inside = true
				break
			}
		}
		if !inside {
			kept = append(kept, loc)
		}
	}
	return kept
}
