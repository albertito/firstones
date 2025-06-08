package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

//go:embed http/*
var httpFS embed.FS

var rootTemplate *template.Template

func serveHTTP(addr string) {
	log.SetFlags(log.Lshortfile)

	go signalHandler()

	tmplFuncs := template.FuncMap{
		"join": func(sep string, a []string) string {
			return strings.Join(a, sep)
		},
	}

	rootTemplate = template.Must(
		template.New("root").Funcs(tmplFuncs).
			ParseFS(httpFS, "http/index.tmpl.html"))

	http.HandleFunc("GET /{$}", handleRoot)
	http.HandleFunc("GET /svg", handleSVG)
	http.HandleFunc("PUT /svg", handleSVG)

	log.Printf("firstones %s", Version())
	log.Printf("Starting HTTP server on %q", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func signalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	s := <-c
	fatalf("Received signal %v, shutting down", s)
}

func wordsFromRequest(r *http.Request) []string {
	r.ParseForm()
	wordsF := r.Form["words"]

	// For extra convenience, the words can be provided as a single
	// space-separated string.
	words := []string{}
	for _, wordF := range wordsF {
		words = append(words, strings.Fields(wordF)...)
	}

	// Max number of words supported is 10.
	if len(words) > 10 {
		words = words[:10]
	}

	return words
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	words := wordsFromRequest(r)

	svg := ""
	var svgErr error
	if len(words) > 0 {
		svg, svgErr = genSVG(words, r.FormValue("grid") == "1")
	}

	data := map[string]interface{}{
		"Words": words,

		// The generated HTML should be already safe for embedding.
		"SVG":   template.HTML(svg),
		"Error": svgErr,
	}
	err := rootTemplate.ExecuteTemplate(w, "index.tmpl.html", data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err),
			http.StatusInternalServerError)
		return
	}
}

func genSVG(words []string, grid bool) (string, error) {
	wsvg, width, height, err := wordsToSVG(words)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	buf.WriteString(string(svgHeader(width, height)))

	writeDefs(buf)

	if grid {
		buf.WriteString(string(svgGrid(width, height)))
	}

	buf.WriteString(string(wsvg))
	buf.WriteString("</svg>\n")

	return buf.String(), nil
}

func handleSVG(w http.ResponseWriter, r *http.Request) {
	words := wordsFromRequest(r)
	if len(words) == 0 {
		http.Error(w, "No words provided", http.StatusBadRequest)
		return
	}

	svg, err := genSVG(words, r.FormValue("grid") == "1")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating SVG: %v", err),
			http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write([]byte(svg))
}
