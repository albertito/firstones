package main

import (
	"bufio"
	"embed"
	"encoding/xml"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// # Glyph definitions
//
// Each glyph is stored in a separate SVG file in the "glyphs" directory.
// The name of the file is the name of the glyph, e.g. "fEEt.svg" for the
// "fEEt" glyph.
// The names of the glyphs follow the official PDF.
//
// # Glyph geometry
//
// On the X axis, glyphs are centered on X=0 (so they have some parts on
// X<0). They have different widths, but the max is 16.
// On the Y axis, glyphs _begin_ at Y=0, and flow downwards (so they don't
// have anything on Y<0). They have different heights, but the max is 16.
//
// TODO: review the min/max, maybe we want easier numbers?
//
// That allows us to assume that the initial connection point for all glyphs
// is on their (0,0). And the end point is on (0, $height).

//go:embed glyphs/*.svg
var glyphsFS embed.FS

type Glyph struct {
	name      string // e.g. "fEEt"
	svgDef    SVG    // The full SVG definition.
	svg       SVG    // The SVG referencing this glyph.
	height    int
	connector bool
}

type Syllable []Glyph

func (s Syllable) String() string {
	names := make([]string, 0, len(s))
	for _, g := range s {
		names = append(names, g.name)
	}
	return strings.Join(names, "-")
}

type Word []Syllable

func (word Word) String() string {
	syS := make([]string, 0, len(word))
	for _, syllable := range word {
		syS = append(syS, syllable.String())
	}
	return strings.Join(syS, "/")
}

func (g Glyph) String() string {
	return "[" + g.name + "]"
}

var allGlyphs = map[string]Glyph{}

// Sorted list of glyph names.
var glyphNames = []string{}

func getGlyph(name string) (Glyph, error) {
	g, ok := allGlyphs[name]
	if !ok {
		return g, fmt.Errorf("Unknown glyph %q", name)
	}
	return g, nil
}

func mustGetGlyph(name string) Glyph {
	g, err := getGlyph(name)
	if err != nil {
		panic(err)
	}
	return g
}

// phonemesToGlyphs converts a slice of phonemes into a slice of Glyphs.
// Phonemes is a string with the individual phonemes separated by "-".
// The "/" phoneme is used to indicate a new syllable.
// This returns a slice of syllables (as Glyphs).
func phonemesToGlyphs(phonemes string) (Word, error) {
	word := Word{}

	// We want to support a variety of ways to handle the "/" end-of-syllable
	// marker:
	//   - SH-fEEt-/-R-All
	//   - SH-fEEt/R-All
	//   - SH-fEEt/-R-All
	//   - SH-fEEt-/R-All
	//
	// Also cases like T//T or T/-/T should be handled well.
	for _, syllableS := range strings.Split(phonemes, "/") {
		syllable := Syllable{}
		for _, phoneme := range strings.Split(syllableS, "-") {
			if phoneme == "" {
				// This can happen on cases like "T-/-T", or "T//T".
				continue
			}

			g, err := getGlyph(phoneme)
			if err != nil {
				return nil, err
			}
			syllable = append(syllable, g)
		}

		if len(syllable) > 0 {
			word = append(word, syllable)
		}
	}

	return word, nil
}

func init() {
	// We load all the glyphs from the embedded filesystem.
	des, err := glyphsFS.ReadDir("glyphs")
	if err != nil {
		panic(err)
	}

	for _, de := range des {
		name := strings.TrimSuffix(de.Name(), ".svg")
		content, err := glyphsFS.ReadFile("glyphs/" + de.Name())
		if err != nil {
			panic(err)
		}

		// We extract some information from the SVG definition itself.
		var (
			// The element ID.
			id     string
			height int  // Height, stored in _fo_height.
			conn   bool // Is this a connector? In _fo_connector.
		)

		firstElem, err := extractFirstElement(string(content))
		if err != nil {
			panicf("%s extractFirstElement: %v", de.Name(), err)
		}
		for _, attr := range firstElem.Attr {
			switch attr.Name.Local {
			case "id":
				id = strings.TrimPrefix(attr.Value, "glyph:")
				if id != name {
					panicf("%s id does not match name '%s'", de.Name(), id)
				}
			case "_fo_height":
				height, err = strconv.Atoi(attr.Value)
				if err != nil {
					panicf("%s _fo_height: %v", de.Name(), err)
				}
			case "_fo_connector":
				conn, err = strconv.ParseBool(attr.Value)
				if err != nil {
					panicf("%s _fo_connector: %v", de.Name(), err)
				}

			}
		}

		allGlyphs[name] = Glyph{
			name:   name,
			svgDef: SVG(string(content)),
			svg: SVGf(
				`<use href="#glyph:%s" />`, name),
			height:    height,
			connector: conn,
		}
	}

	glyphNames = slices.Sorted(maps.Keys(allGlyphs))
}

func extractFirstElement(svgDef string) (xml.StartElement, error) {
	dec := xml.NewDecoder(strings.NewReader(svgDef))
	tok, err := dec.Token()
	return tok.(xml.StartElement), err
}

func dumpGlyphs(w io.Writer) {
	buf := bufio.NewWriter(w)
	buf.WriteString(string(svgHeader(80, 210)))
	buf.WriteString(string(svgGrid(80, 210)))
	writeDefs(buf)

	// Start at (10, 10) and move through the glyphs row by row.
	x, y := 10, 10

	// Names taken from the official PDF, sorted as they appear there.
	// "_" is used to put a new line in the SVG, to match the official PDF.
	names := []string{
		"B", "CH", "D", "DH", "_",
		"F", "G", "H", "J", "_",
		"K", "L", "M", "N", "_",
		"NG", "P", "R", "S", "_",
		"SH", "T", "TH", "V", "_",
		"W", "Z", "ZH", "_",
		"sAd", "All", "sAy", "_",
		"pEt", "fEEt", "lIt", "I", "_",
		"gOOd", "tOO", "gO", "_",
		"hOUse", "fUn", "bOY", "Yes",
	}

	for _, name := range names {
		if name == "_" {
			// Move to the next row, print a horizontal line to separate.
			x = 10
			y += 20
			fmt.Fprintf(buf,
				`<line x1="0" y1="%d" x2="100" y2="%d" `+
					`stroke="black" stroke-width="0.5" />`+"\n",
				y-6, y-6)
			continue
		}

		g := mustGetGlyph(name)

		s := SVGfn(
			`<text x="-2" y="-3" font-size="2" fill="black" `+
				`font-family="sans-serif">%s`,
			g.name)
		conn := ""
		if g.connector {
			conn = ", 🔗"
		}
		s += SVGfn(`<tspan font-size="1.5">(%d%s)</tspan>`, g.height, conn)
		s += SVGfn(`</text>`)

		// The glyph.
		s += color("orange", g.svg)
		s = move(x, y, s)
		buf.WriteString(string(s) + "\n")

		// Little dot marking the top of the glyph, to validate shape.
		fmt.Fprintf(buf,
			`<circle cx="%d" cy="%d" r="0.2" `+
				`fill="darkorange" />`+"\n",
			x, y)

		// Little dot marking the bottom of the glyph, to validate height.
		fmt.Fprintf(buf,
			`<circle cx="%d" cy="%d" r="0.2" `+
				`fill="darkorange" />`+"\n",
			x, y+g.height)

		x += syllableSpacing
	}

	buf.WriteString("</svg>\n")
	buf.Flush()
}
