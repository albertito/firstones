package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestMapSyllables(t *testing.T) {
	// Glyphs for convenience.
	g1, g2 := mustGetGlyph("SH"), mustGetGlyph("fEEt")
	g3, g4 := mustGetGlyph("R"), mustGetGlyph("All")
	g5, g6 := mustGetGlyph("S"), mustGetGlyph("hOUse")

	cases := []struct {
		glyphs   []Glyph
		slashes  []int
		wlen     int
		expected Word
	}{
		{
			// No slashes.
			glyphs:   []Glyph{g1, g2, g3, g4, g5, g6},
			slashes:  []int{},
			wlen:     8,
			expected: Word{{g1, g2, g3, g4, g5, g6}},
		},
		// Same length as the original word, all possible location for a
		// single slash.
		{
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{0},
			wlen:     4,
			expected: Word{{g1}, {g2, g3, g4}},
		},
		{
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{1},
			wlen:     4,
			expected: Word{{g1}, {g2, g3, g4}},
		},
		{
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{2},
			wlen:     4,
			expected: Word{{g1, g2}, {g3, g4}},
		},
		{
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{3},
			wlen:     4,
			expected: Word{{g1, g2, g3}, {g4}},
		},
		{
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{4},
			wlen:     4,
			expected: Word{{g1, g2, g3}, {g4}},
		},
		{
			// Longer word, with a slash > len(glyphs).
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{6},
			wlen:     8,
			expected: Word{{g1, g2, g3}, {g4}},
		},
		{
			// Slash at 0. This is a special case, where the first syllable is
			// the first glyph.
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{0},
			wlen:     8,
			expected: Word{{g1}, {g2, g3, g4}},
		},
		{
			// Slash at wlen. This is a special case, where the last syllable
			// is the last glyph.
			glyphs:   []Glyph{g1, g2, g3, g4},
			slashes:  []int{8},
			wlen:     8,
			expected: Word{{g1, g2, g3}, {g4}},
		},
		{
			// Multiple slashes.
			glyphs:   []Glyph{g1, g2, g3, g4, g5, g6},
			slashes:  []int{2, 4},
			wlen:     6,
			expected: Word{{g1, g2}, {g3, g4}, {g5, g6}},
		},
		{
			// Multiple slashes in the exact same place.
			glyphs:   []Glyph{g1, g2, g3, g4, g5, g6},
			slashes:  []int{2, 2, 4, 4},
			wlen:     6,
			expected: Word{{g1, g2}, {g3, g4}, {g5, g6}},
		},
		{
			// Multiple slashes, with wlen != len(glyphs).
			glyphs:   []Glyph{g1, g2, g3, g4, g5, g6},
			slashes:  []int{6, 7},
			wlen:     12,
			expected: Word{{g1, g2, g3}, {g4, g5, g6}},
		},
	}
	for i, c := range cases {
		result := mapSyllables(c.glyphs, c.slashes, c.wlen)
		t.Logf("%d: %v %v %d -> %v", i, c.glyphs, c.slashes, c.wlen, result)

		diff := cmp.Diff(c.expected, result,
			cmpopts.EquateComparable(Glyph{}))
		if diff != "" {
			t.Errorf("%d: Mismatch in syllables:\n%s", i, diff)
		}
	}
}
