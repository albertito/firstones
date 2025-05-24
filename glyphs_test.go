package main

import "testing"

func mkS(names ...string) Syllable {
	s := make(Syllable, len(names))
	for i, name := range names {
		s[i] = allGlyphs[name]
	}
	return s
}

func TestPhonemesToGlyphs(t *testing.T) {
	type Case struct {
		ps  string
		w   Word
		err bool
	}
	cases := []Case{
		{
			ps: "SH-fEEt-R-All",
			w:  Word{mkS("SH", "fEEt", "R", "All")},
		},
		{
			ps:  "SH-fE/Et-All",
			err: true,
		},
		{
			ps: "lIt-N/T-R-sAd-P/T-All",
			w: Word{
				mkS("lIt", "N"),
				mkS("T", "R", "sAd", "P"),
				mkS("T", "All")},
		},
	}

	// Variants for the same output.
	for _, s := range []string{
		"SH-fEEt-/-R-All", "/-SH-fEEt-/-R-All-/",
		"-SH-fEEt-/-R-All-", "SH-fEEt-/-/-R-All",
		"SH-fEEt/R-All", "SH-fEEt//R-All",
		"/SH-fEEt//R-All/",
	} {
		cases = append(cases, Case{
			ps: s,
			w:  Word{mkS("SH", "fEEt"), mkS("R", "All")},
		})
	}

	for i, c := range cases {
		got, err := phonemesToGlyphs(c.ps)
		t.Logf("%d: phonemesToGlyphs(%q) = %v / %v, want %v / %v",
			i, c.ps, got, err, c.w, c.err)
		if (err == nil) != !c.err {
			t.Errorf("   error mismatch: got %v, want %v", err, c.err)
		}
		if len(got) != len(c.w) {
			t.Errorf("   length mismatch: got %d, want %d", len(got), len(c.w))
		}
	}
}

func TestMustGetGlyphErr(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("mustGetGlyph did not panic on unknown glyph")
		}
	}()

	mustGetGlyph("unknown-glyph")
}
