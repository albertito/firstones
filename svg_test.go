package main

import "testing"

func TestSVGf(t *testing.T) {
	type Case struct {
		fmt  string
		args []interface{}
	}

	safe := []Case{
		{"<x>", nil},
		{"<x></x>", nil},
		{"<x>%s</x>", []interface{}{""}},
		{"<x>%s</x>", []interface{}{"hola"}},
		{"<x>%v</x>", []interface{}{SVG("hola")}},
		{"<x>%v</x>", []interface{}{3}},
	}
	for _, c := range safe {
		// Just call the function, we only want to check it doesn't panic.
		// Correctness of output is well covered in the integration tests.
		SVGf(c.fmt, c.args...)
	}

	// Test cases for detecting unsafe characters in SVG strings.
	unsafe := []Case{
		{"<x>%s</x>", []interface{}{"<"}},
		{"<x>%s</x>", []interface{}{">"}},
		{"<x>%s</x>", []interface{}{"<y>"}},
	}
	for i, c := range unsafe {
		expectSVGfPanic(t, i, c.fmt, c.args)
	}
}

func expectSVGfPanic(t *testing.T, i int, f string, args []interface{}) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%d: expected panic - fmt %q, args %v", i, f, args)
		}
	}()
	SVGf(f, args...)
}
