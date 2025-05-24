package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

const usage = `# firstones - convert words to She-Ra First Ones language

Usage:

  firstones [flags] svg [words...]
    Generate an SVG image with the given words, printed to stdout.
  firstones [flags] http <address>
    Start a web server at the given address.
  firstones [flags] dump-glyphs
    Generate an SVG image with all the glyphs, for debugging.
  firstones [flags] version
    Print software version information.

Flags:
`

var (
	showGrid = flag.Bool("grid", false,
		"show grid in the svg, for debugging")
)

func Usage() {
	fmt.Fprintf(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintln(os.Stderr)
	os.Exit(1)
}

func panicf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func Version() string {
	info, _ := debug.ReadBuildInfo()
	rev := info.Main.Version
	ts := time.Time{}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			ts, _ = time.Parse(time.RFC3339, s.Value)
		}
	}
	return fmt.Sprintf("version %s (%s)", rev, ts)
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	switch flag.Arg(0) {
	case "version":
		fmt.Println(Version())
		os.Exit(0)
	case "dump-glyphs":
		dumpGlyphs(os.Stdout)
	case "svg":
		words := []string{}
		args := flag.Args()
		for i := 1; i < len(args); i++ {
			for _, w := range strings.Fields(args[i]) {
				words = append(words, w)
			}
		}
		if len(words) == 0 {
			words = []string{"SH-fEEt-R-All"}
		}
		printSVG(words)
	case "http":
		if len(flag.Args()) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: firstones http <address>")
			os.Exit(1)
		}
		address := flag.Args()[1]
		serveHTTP(address)
	default:
		Usage()
		os.Exit(1)
	}
}

func printSVG(words []string) {
	wsvg, width, height, err := wordsToSVG(words)
	if err != nil {
		fatalf("error converting words to SVG: %v", err)
	}

	fmt.Print(string(svgHeader(width, height)))
	writeDefs(os.Stdout)

	if *showGrid {
		fmt.Print(string(svgGrid(width, height)))
	}

	fmt.Print(string(wsvg))
	fmt.Println("</svg>")
}
