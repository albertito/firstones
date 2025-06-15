package main

import (
	"fmt"
	"io"
	"math"
	"strings"
)

// SVG type, to make it easier to avoid mixing it up with normal strings.
type SVG string

func SVGf(format string, args ...interface{}) SVG {
	out := fmt.Sprintf(format, args...)

	// Paranoid check: we should not have any more '<' or '>' than
	// the ones in the format string.
	// This should be the case because we already filter what goes into the
	// args, but it is better to be safe.
	for _, c := range []string{"<", ">"} {
		fq := strings.Count(format, c)
		oq := strings.Count(out, c)
		if fq != oq {
			panicf("SVGf: unsafe for %q: format %q had %d, output %q had %d",
				c, format, fq, out, oq)
		}
	}

	return SVG(out)
}

const svgNL = SVG("\n")

func SVGfn(format string, args ...interface{}) SVG {
	return SVGf(format, args...) + svgNL
}

// Indent the SVG by n spaces, for readability.
func indent(svg SVG, n int) SVG {
	sp := strings.Repeat(" ", n)
	lines := strings.Split(string(svg), "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = sp + line
		}
	}
	return SVG(strings.Join(lines, "\n"))
}

func move(dx, dy int, svg SVG) SVG {
	return SVGfn(`<g transform="translate(%d %d)">`, dx, dy) +
		indent(svg, 2) + SVG("</g>\n")
}

func movef(dx, dy float64, svg SVG) SVG {
	return SVGfn(`<g transform="translate(%g %g)">`, dx, dy) +
		indent(svg, 2) + SVG("</g>\n")
}

func rotate(angle int, svg SVG) SVG {
	return SVGfn(`<g transform="rotate(%d)">`, angle) +
		indent(svg, 2) + SVG("</g>\n")
}

func color(color string, svg SVG) SVG {
	return SVGfn(
		`<g color="%s" stroke="%s" stroke-width="0.5">`, color, color) +
		indent(svg, 2) + SVG("</g>\n")
}

func vertLine(l int) SVG {
	return SVGf(
		`<line x1="0" y1="0" x2="0" y2="%d" />`, l)
}

func svgHeader(width, height int) SVG {
	return SVGfn(`<svg
  version="1.1"
  viewBox="0 0 %d %d"
  width="%dmm" height="%dmm"
  xmlns="http://www.w3.org/2000/svg">
`, width, height, width, height)
}

func svgGrid(width, height int) SVG {
	s := "<!-- Grid for debugging -->\n"
	s += `<g stroke-width="0.1">` + "\n"
	for i := 0; i <= width; i += 5 {
		stroke := "#eee"
		if i%10 == 0 {
			stroke = "#ccc"
		}
		s += fmt.Sprintf(
			`<line x1="%d" y1="0" x2="%d" y2="100%%" stroke="%s"/> `,
			i, i, stroke)
		s += "\n"
	}
	for i := 0; i <= height; i += 5 {
		stroke := "#eee"
		if i%10 == 0 {
			stroke = "#ccc"
		}
		s += fmt.Sprintf(
			`<line x1="0" y1="%d" x2="100%%" y2="%d" stroke="%s"/>`, i, i, stroke)
	}
	s += "</g> <!-- End of grid -->\n"
	return SVG(s)
}

// Write all the known glyphs in a <defs> section, so they can be referred to
// individually. This makes the SVG more readable.
func writeDefs(w io.Writer) {
	fmt.Fprintln(w, `<defs>`)
	// We write in alphabetical order, so that the SVG output is reproducible.
	for _, name := range glyphNames {
		g := allGlyphs[name]
		fmt.Fprintf(w, "<!-- %s -->\n", name)
		fmt.Fprintln(w, string(g.svgDef))
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w, `</defs>`)
}

func syllableToSVG(syllable Syllable) SVG {
	svg := SVGf("<g> <!-- Syllable: %v -->\n", syllable)
	height := 0

	// Was the previous glyph a connector?
	prevConnector := false
	for _, glyph := range syllable {
		if !glyph.connector && !prevConnector {
			// If the glyph is not a connector, and the previous one was not a
			// connector either, we need to draw a vertical line to connect it
			// to the previous glyph (or the word branch).
			svg += color("orange",
				move(0, height,
					vertLine(3)),
			)
			height += 3
		}

		svg += color("orange",
			move(0, height,
				glyph.svg),
		)
		svg += svgNL

		height += glyph.height
		prevConnector = glyph.connector
	}

	svg += SVGf("</g> <!-- End of syllable %v -->\n", syllable)
	return svg
}

const (
	// How much space we leave between each syllable?
	// This has to be > than the maximum width of a glyph.
	syllableSpacing = 20

	// How much space we leave between each word?
	wordSpacing = 10

	// How much space we leave at the top?
	// This is so elements on the border don't end up getting chopped.
	topMargin = 5
)

type WordLine struct {
	nsyllables int
	angle      int // Angle in degrees.
	svg        SVG
}

// offsetFor returns the X and Y offsets for the i-th syllable in the word line.
func (wl WordLine) offsetFor(i int) (x, y float64) {
	// Where in the line should the i-th syllable start?
	// We divide the line into nsyllables+1 segments.
	totalLength := wl.nsyllables * syllableSpacing
	segmentLength := float64(totalLength) / float64(wl.nsyllables+1)

	// Angle in radiants.
	rad := float64(wl.angle) * math.Pi / 180

	// X and Y offsets for the i-th syllable.
	x = -math.Cos(rad) * segmentLength * float64(i+1)
	y = -math.Sin(rad) * segmentLength * float64(i+1)
	return
}

func (wl WordLine) LenX() float64 {
	// The word line begins at X=0 and ends at
	// X=-(nsyllables * syllableSpacing) BUT we have to account for the angle.
	totalLength := wl.nsyllables * syllableSpacing
	rad := float64(wl.angle) * math.Pi / 180
	return math.Cos(rad) * float64(totalLength)
}

func wordLineSVG(n int) WordLine {
	// Draw the slanted line for the word branch.
	// n is how many syllables we will have, and determines the length of
	// the line.
	wl := WordLine{
		nsyllables: n,

		// Angles in published media:
		//  - Official PDF: 23°, 29.5°, 24.5°, 24°
		//  - "Happy new year": 30°, 27.5°
		//  - "April fools": 23°
		angle: -12,
	}

	wl.svg = SVG("<g> <!-- Word line -->\n")

	// The line sits at Y=0.
	// X goes from -(n * syllableSpacing) to 0: because we draw the text
	// backwards, this makes it easier to position the line.

	// Line.
	startX := -n * syllableSpacing
	s := SVGfn(`<line x1="%d" y1="0" x2="0" y2="0" />`, startX)

	// Dots indicating start and end.
	s += SVGfn(`<circle cx="%d" cy="0" r="0.5" fill="currentcolor" />`, startX)
	s += SVGfn(`<circle cx="0" cy="0" r="0.5" fill="currentcolor" />`)

	wl.svg += rotate(wl.angle, s)

	wl.svg += SVG("</g> <!-- End of word line -->\n")
	return wl
}

func wordsWidthHeight(words []Word) (int, int) {
	// Calculate how wide and tall the words will be in the SVG.
	// Doesn't have to be super accurate (and isn't due to the slanting), it's
	// used to size the general canvas, and compute the starting position.
	width := 0
	height := 0
	maxSyllables := 0
	for _, word := range words {
		if len(word) > maxSyllables {
			maxSyllables = len(word)
		}

		width += syllableSpacing * len(word)
		width += wordSpacing

		for _, syllable := range word {
			sh := 0
			prevC := false
			for _, glyph := range syllable {
				sh += glyph.height
				if !glyph.connector && !prevC {
					// Connector line.
					sh += 3
				}
				prevC = glyph.connector
			}
			if sh > height {
				height = sh
			}
		}
	}

	// We start at Y=topMargin, so we add that to the height.
	height += topMargin

	// Leave some margin for the rotation.  This is not accurate, but works
	// for reasonable syllable lengths.
	height += 5 * maxSyllables

	return width, height
}

func wordsToSVG(words []string) (SVG, int, int, error) {
	for _, word := range words {
		// The word is user-provided. Check it doesn't contain any problematic
		// characters that would cause issues in the SVG.
		if err := isSafeForSVG(word); err != nil {
			return SVG(""), 0, 0, fmt.Errorf(
				"word %q is not safe for SVG: %v", word, err)
		}
	}

	svg := SVGfn("<!-- Words: %v -->", words)

	// wordsG contains the words, as syllables of Glyphs.
	wordsG := []Word{}
	for _, word := range words {
		wordG, err := smartWordToGlyphs(word)
		if err != nil {
			return svg, 0, 0, fmt.Errorf(
				"error converting %q to glyphs: %v", word, err)
		}

		if len(wordG) == 0 {
			// Skip empty words, this can happen with words that contain just
			// "-" or "/".
			continue
		}

		wordsG = append(wordsG, wordG)
	}

	// The language is right to left, so we compute the total width, and start
	// there (+ some margin) and go backwards.
	width, height := wordsWidthHeight(wordsG)
	x := float64(width)

	for _, wordG := range wordsG {
		svg += SVGfn("<!-- Glyphs for %v -->", wordG)

		wl := wordLineSVG(len(wordG))
		wsvg := wl.svg
		for i, syllable := range wordG {
			offx, offy := wl.offsetFor(i)
			wsvg += movef(offx, offy,
				syllableToSVG(syllable))
		}

		svg += movef(x, topMargin,
			color("orange", wsvg))

		x -= wl.LenX()
		x -= wordSpacing
	}

	return svg, width + wordSpacing, height, nil
}

func isSafeForSVG(word string) error {
	// For our use cases, we don't expect any punctuation characters other
	// than '/' to separate syllables.
	// And in particular we don't want to allow any characters that could
	// alter the SVG structure, like '<' or '>'.
	// Note that are characters that still could cause errors later on, but
	// they wouldn't be problematic to include in an SVG.
	for _, r := range word {
		if r == '<' || r == '>' || r == '&' ||
			r == '"' || r == '\\' {
			return fmt.Errorf("unsafe character %q", r)
		}
	}
	return nil
}
