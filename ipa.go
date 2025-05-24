package main

import (
	"bufio"
	"embed"
	"fmt"
	"strings"

	"golang.org/x/text/language"
)

// We embed the IPA dictionaries for convenience, and parse them internally.
// They come from the open-dict-data project, and we use the tab-delimited
// text format for convenience.

//go:embed ipa/*.txt
var ipaFS embed.FS

// Map of word -> pronunciation (as IPA symbols).
type IPADict map[string]string

// The known dictionaries.
var IPADicts = map[int]IPADict{}

// Language matcher, to find the correct IPA dictionary.
var langMatcher language.Matcher

// Map the names of the She-Ra characters to their IPA.
var namesIPA = map[string]string{
	"she-ra":      "ʃiɹɑ",
	"shera":       "ʃiɹɑ",
	"catra":       "kætɹa",
	"entrapta":    "ɪntɹæpta",
	"hordak":      "hɔɹdak",
	"perfuma":     "pɝfjuma",
	"scorpia":     "skɔɹpia",
	"mermista":    "mɝmistə",
	"frosta":      "fɹɔsta",
	"netossa":     "nɛttɔsa",
	"spinnerella": "spɪnɝɛɫə",
	"etheria":     "iθiɹɑ",

	// These are words in English, but we want to make sure they're always
	// identified.
	"adora":   "ədɔɹə",
	"glimmer": "ɡɫɪmɝ",
	"bow":     "boʊ",
	"shadow":  "ʃædoʊ", "weaver": "wivɝ",
	"swift": "swɪft", "wind": "wɪnd",
	"sea": "si", "hawk": "hɔk",
}

func init() {
	// We load all the IPA dictionaries from the embedded filesystem.
	// This is wildly inefficient if we only use it once, but considering
	// the size of the dictionaries, it's not a problem, and we are optimizing
	// for repeated use.
	des, err := ipaFS.ReadDir("ipa")
	if err != nil {
		panic(err)
	}
	langs := []language.Tag{}
	for i, de := range des {
		langS := strings.TrimSuffix(de.Name(), ".txt")
		lang := language.MustParse(langS)
		dict := IPADict{}
		langs = append(langs, lang)

		// The order in which we add it to langs identifies this dictionary.
		// The matcher will return this index when doing a match.
		IPADicts[i] = dict

		// Scan line by line. Format is:
		//   word<TAB>/pronunciation1/, /pronunciation2/, ...
		f, err := ipaFS.Open("ipa/" + de.Name())
		if err != nil {
			panic(err)
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			word, prons, ok := strings.Cut(line, "\t")
			if !ok || word == "" || prons == "" {
				continue
			}

			// We keep the first pronunciation, as for our use case we only
			// need one.
			slspl := strings.Split(prons, "/")
			if len(slspl) < 2 {
				continue
			}
			// TODO: clean up the symbols, remove accents, etc.
			dict[word] = strings.TrimSpace(slspl[1])
		}

		// Add the namesIPA to all dictionaries.
		// This also overrides the word if it exists. That's okay.
		for name, ipa := range namesIPA {
			dict[name] = ipa
		}
	}

	langMatcher = language.NewMatcher(langs, language.PreferSameScript(true))
}

// IPA symbol to first ones glyph mapping.
// This is manually curated, and we do our best to match symbols to glyphs.
// There are gaps which are filled in by approximation.
// See ipa/symbols.py for the helper used to extract the list from the
// dictionaries.
//
// To approximate and confirm the mappings, we use the following sources as
// starting points:
// - https://en.wikipedia.org/wiki/Pronunciation_respelling_for_English
// - https://en.wikipedia.org/wiki/IPA_consonant_chart_with_audio
// - https://en.wikipedia.org/wiki/IPA_vowel_chart_with_audio
// - The official PDF which contains some examples.
// - Official "Happy new year" and "April fool" published sigils.
//
// Some of the mappings were also confirmed by cross-checking with
// enby_lydia@discord's hand-made IPA map, which they posted on 2025-05-27 on
// #member-fan-content in the She-Ra discord server.

// We have two maps: ipaToGlyphs2 maps 2-symbol IPA sequences to glyphs,
// and ipaToGlyphs1 maps single-symbol IPA sequences to glyphs.

var ipaToGlyphs2 = map[string]string{
	"aʊ": "hOUse", // "house" en_US lookup
	"ɔɪ": "bOY",   // "boy" en_US lookup
	"oʊ": "gO",    // "go" en_US lookup
	"tʃ": "CH",    // Wikipedia (e.g. "CHurCH").
	"dʒ": "J",     // Wikipedia (e.g.: "Jump").
	"aɪ": "I",     // "I", "bY" en_US lookup
}

var ipaToGlyphs1 = map[rune]string{
	//
	// Consonants
	// Sorted by the value, in the order they appear in the official PDF.
	//
	'b': "B", // Wikipedia.
	'β': "B", // Closest match. Does not appear in en_US. es_MX: "beBiBle".
	// CH is in ipaToGlyphs2 (tʃ).
	'd': "D",  // "Shadow weaver" official PDF, Wikipedia.
	'ð': "DH", // Wikipedia. Example: "THis".

	'f': "F", // Wikipedia.
	'g': "G", // Wikipedia.
	'ɡ': "G", // Another IPA symbol for 'g' (historic, see Wikipedia for IPA).
	'ɣ': "G", // Closest match. Does not appear in en_US. es_MX: "borreGo".
	'h': "H", // Wikipedia.
	// J is in ipaToGlyphs2 (dʒ).

	'k': "K", // "Scorpia" official PDF -> "scorpio" IPA.
	'x': "K", // Closest match. es_MX: "Jota". Not very close :(
	'l': "L", // Wikipedia.
	'ɫ': "L", // "April Fool" official sigil -> "april" IPA.
	'ʎ': "L", // Closest match. Does not appear in en_US. es_MX: "LLuvia".
	'ʝ': "L", // Closest match. Does not appear in en_US. es_MX: "aLLa".
	'm': "M", // Wikipedia.
	'n': "N", // "Happy new year" official sigil -> "new" IPA.
	'ɲ': "N", // Closest match. Does not appear in en_US, it's Ñ in es.

	'ŋ': "NG", // Wikipedia.
	'p': "P",  // "April Fool" official sigil -> "april" IPA.
	'ɹ': "R",  // "April Fool" -> "april" IPA
	'ɾ': "R",  // Closest match for this symbol.
	'r': "R",  // Closest match, this is a hard R (e.g. "feRRocaRRil").
	'ɝ': "R",  // Closest match, from wikipedia (R-colored_vowel) -> assERt
	's': "S",  // "Scorpia" official PDF -> "scorpio" IPA.

	'ʃ': "SH", // "She-ra" official PDF -> "she" IPA.
	't': "T",  // Wikipedia.
	'θ': "TH", // Wikipedia.
	'v': "V",  // "Shadow weaver" official PDF -> "weaver" IPA.

	'w': "W",  // "Shadow weaver" official PDF -> "weaver" IPA
	'z': "Z",  // Wikipedia.
	'ʒ': "ZH", // Wikipedia (e.g. "vision")

	//
	// Vowels
	// Sorted and grouped as they appear in the official PDF.
	//
	'æ': "sAd", // "sad" en_US lookup
	'a': "sAd", // IPA vowels chart: "æ" is the closest to "a".
	'ɔ': "All", // "all" en_US lookup
	'ɑ': "All", // "ra" IPA -> "ɹɑ", "She-Ra" official PDF.
	'o': "All", // IPA vowels chart: "ɔ" is the closest to "o".
	'e': "sAy", // "say" en_US lookup

	'ɛ': "pEt",  // "pet" en_US lookup
	'i': "fEEt", // "feet" en_US lookup
	'ɪ': "lIt",  // "lit" en_US lookup
	// "I" is in ipaToGlyphs2 (aɪ).

	'ʊ': "gOOd", // "good" en_US lookup
	'u': "tOO",  // "too" en_US lookup
	// "gO" is in ipaToGlyphs2.

	// "hOUse" is in ipaToGlyphs2.
	'ə': "fUn", // "fun" en_US lookup
	// "bOY" is in ipaToGlyphs2.
	'j': "Yes", // "yes" en_US lookup

	//
	// Ignored
	//
	'ˈ': "", // Primary stress mark.
	'ˌ': "", // Secondary stress mark.
}

// langWord ToGlyphs converts a word in the given language, to a glyph
// Word.
func langWordToGlyphs(word, lang string) (Word, error) {
	langTag, langIdx, confidence := langMatcher.Match(language.Make(lang))
	if confidence <= language.Low {
		return nil, fmt.Errorf("language not supported (sorry!)")
	}

	dict, ok := IPADicts[langIdx]
	if !ok {
		return nil, fmt.Errorf("no dictionary for language %q", langTag)
	}

	// We use a heuristic for the syllables.
	// A '/' in the input indicates a new syllable. We record where they are
	// in the input, then try to match them on the output.
	syllablesIdxs := findSlashes(word)
	word = strings.ReplaceAll(word, "/", "")

	// Get the IPA representation of the word.
	ipa, ok := dict[word]
	if !ok {
		// Try the lowercase variant, for convenience.
		// Note we can't just lowercase because some IPA dicts have
		// intentional uppercase words.
		ipa, ok = dict[strings.ToLower(word)]
		if !ok {
			return nil, fmt.Errorf("unknown word")
		}
	}

	// The conversion of IPA representation to glyphs is annoying, because we
	// have to account for the two-symbol sequences.
	ipaR := []rune(ipa)
	glyphs := []Glyph{}
	for i := 0; i < len(ipaR); i++ {
		// Look up this and the next rune in the two-symbol map.
		// If we have a match, use it and skip the next rune.
		if i+1 < len(ipaR) {
			s := string(ipaR[i]) + string(ipaR[i+1])
			if glyph, ok := ipaToGlyphs2[s]; ok {
				glyphs = append(glyphs, mustGetGlyph(glyph))
				i++
				continue
			}
		}

		// No match in the two-symbol map, look up this single symbol.
		if glyph, ok := ipaToGlyphs1[ipaR[i]]; ok {
			if glyph == "" {
				// Intentionally ignore empty glyphs.
				continue
			}
			glyphs = append(glyphs, mustGetGlyph(glyph))
		} else {
			return nil, fmt.Errorf("unknown IPA symbol %q", ipaR[i])
		}
	}

	return mapSyllables(glyphs, syllablesIdxs, len(word)), nil
}

// findSlashes finds the indices of the slashes in the word.
func findSlashes(word string) []int {
	slashes := []int{}
	for i, r := range word {
		if r == '/' {
			slashes = append(slashes, i)
		}
	}
	return slashes
}

// mapSyllables maps the glyphs to syllables based on the indices of the
// slashes.
func mapSyllables(glyphs []Glyph, slashes []int, wlen int) Word {
	if len(slashes) == 0 {
		// No slashes, just return the whole word as a single syllable.
		return Word{Syllable(glyphs)}
	}

	// The location of the slashes was from the original word; the glyphs
	// have a different length (because of the IPA translation), so we use a
	// horrible heuristic to map them.
	// We map them proportionally to the length of the original word.
	word := Word{}
	prev := 0
	for _, idx := range slashes {
		var gidx int
		if len(glyphs) == wlen {
			// If the glyphs are the same length as the original word,
			// we can just use the indices directly.
			gidx = idx
		} else {
			// Where in the original word was the slash, proportionally.
			f := float64(idx) / float64(wlen)
			// Find the corresponding index in the glyphs.
			gidx = int(float64(len(glyphs)) * f)
		}

		if gidx == 0 {
			// If we map to the beginning of the word, assume we want at least
			// the first glyph.
			gidx = 1
		}
		if gidx >= len(glyphs) {
			// If we map the end of the glyphs, include the last glyph.
			gidx = len(glyphs) - 1
		}
		if gidx <= prev {
			// If we map to the same place we already did, skip it.
			continue
		}

		// Add the syllable from the previous index to the current index.
		word = append(word, glyphs[prev:gidx])
		prev = gidx
	}

	// If we have any glyphs left after the last slash, add them.
	if prev < len(glyphs) {
		word = append(word, glyphs[prev:])
	}

	return word
}

// smartWordToGlyphs converts a string word to a glyph Word.
// If the word has a dictionary prefix (e.g. "en:shadow"), we use that.
// Otherwise, we search through some known IPA dictionaries to try to find it.
// And if that fails, we assume the word is a sequence of phonemes
// (e.g. "SH-fEEt-R-All").
// Syllables are separated by "/", and when doing IPA conversion we do a
// best-effort heuristic mapping.
func smartWordToGlyphs(word string) (Word, error) {
	if lang, w, ok := strings.Cut(word, ":"); ok {
		if lang == "" || lang == "firstones" {
			return phonemesToGlyphs(w)
		}
		// Language-prefixed word.
		return langWordToGlyphs(w, lang)
	}

	// No language prefix, try to find it through some known languages.
	ds := []string{"es", "en"}
	for _, lang := range ds {
		gs, err := langWordToGlyphs(word, lang)
		if err == nil {
			return gs, nil
		}
	}

	// We couldn't find the word so we assume it's a sequence of phonemes.
	return phonemesToGlyphs(word)
}
