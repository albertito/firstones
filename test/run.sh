#!/bin/bash

set -e

. "$(dirname "${0}")/lib.sh"

echo "# Build"
build

echo "# Go tests"
( cd ..; go test )

#
# Rendering tests
#
# We regenerate some SVGs, render them, and do a visual comparison with the
# goldens stored in the repository.
echo "# Rendering tests"

rm -rf .1-svgs/ .2-render-logs .2-render/ .3-compare/ .4-output/ \
	.5-http/ .5-http-logs/

echo "## Generating SVGs"
mkdir -p .1-svgs/
fo dump-glyphs > .1-svgs/dump.svg
fo svg > .1-svgs/default.svg
fo --grid svg > .1-svgs/grid-default.svg
fo --grid svg en/trap/ta > ".1-svgs/grid-en%trap%ta.svg"

if ! ls --zero golden/auto/*.png | parallel -0 --results .1-svgs/{/.}.svg \
	fo svg '{= s:.*/::; s:\.[^/.]+$::; s:%:/:; =}'
then
	cat .1-svgs/*.err
	fail
fi

echo "## Generating PNGs"
mkdir -p .2-render/
if ! ls --zero .1-svgs/*.svg | parallel -0 --results .2-render-logs/{/.} \
	chromium --headless=new --incognito --window-size=400,1000 \
		--screenshot=.2-render/{/.}.png \
		.1-svgs/{/.}.svg ;
then
	cat .2-render-logs/*.err
	fail
fi

echo "## Comparing PNGs to goldens"
if ! ls --zero golden/*.png golden/auto/*.png | parallel -0 --results .3-compare/{/.} \
	magick compare -metric SSIM -compose src \
		{} .2-render/{/.}.png  .2-render/{/.}-diff.png ;
then
	for i in .3-compare/*.err; do
		if grep -q '^1$' "$i"; then
			# The compare command prints "1" on equal files.
			continue
		fi
		C="$(basename "$i" .err)"
		echo "$C differs!"
		echo "   view:   'test/.2-render/$C.png'"
		echo "   diff:   'test/.2-render/$C-diff.png'"
		echo "   accept: cp 'test/.2-render/$C.png' test/golden/auto/"
		echo
	done
	fail
fi


echo "# Command line"
mkdir -p .4-output/
function run_and_compare() {
	fo $@ > ".4-output/$@.stdout" 2> ".4-output/$@.stderr" || true
	./rediff.py "cli/$@.stdout" ".4-output/$@.stdout"
	./rediff.py "cli/$@.stderr" ".4-output/$@.stderr"
}
export -f run_and_compare

if ! ls --zero cli/*.stdout | parallel -0 run_and_compare {/.} ; then
	fail
fi


echo "# HTTP tests"
mkdir -p .5-http/
fo_bg http 127.0.0.1:10294
wait_until_ready 10294

function http_get_and_compare() {
	curl -sS "http://127.0.0.1:10294/$1" > ".5-http/$1"
	CHECKED=0
	if [ -f "http/$1.exact" ]; then
		CHECKED=1
		if ! diff -q "http/$1.exact" ".5-http/$1"; then
			(
				echo "http://127.0.0.1:10294/$1 differs"
				echo
				diff -u "http/$1.exact" ".5-http/$1"
			) > ".5-http/$1.failed"
			exit 1
		fi
	fi
	if [ -f "http/$1.grep" ]; then
		CHECKED=1
		if ! grep -q -f "http/$1.grep" ".5-http/$1"; then
			(
				echo "Grepping content of .5-http/$1 failed"
				echo "  $ grep -f http/$1.grep  .5-http/$1"
				echo "  Grepping for: $(cat "http/$1.grep")"
				echo "  wc .5-http/$1: $(cat ".5-http/$1" | wc)"
			) > ".5-http/$1.failed"
			exit 1
		fi
	fi
	if [ -f "http/$1.jpg" ] ; then
		CHECKED=1
		# Chromium will use the extension to decide how to render it,
		# so we need to rename our file to match.
		case "$(file -b --mime-type ".5-http/$1")" in
		"image/svg+xml")
			EXT="svg"
			;;
		"text/html")
			EXT="html"
			;;
		*)
			(
				echo ".5-http/$1 unknown file type"
				file -b --mime-type ".5-http/$1"
			) > ".5-http/$1.failed"
			exit 1
			;;
		esac
		cp ".5-http/$1" ".5-http/$1.$EXT"
		if ! chromium --headless=new --incognito --window-size=1000,1000 \
			"--screenshot=.5-http/$1.jpg" \
			".5-http/$1.$EXT" ; then
			echo "ERROR rendering .5-http/$1.$EXT" \
				> ".5-http/$1.failed"
			exit 1
		fi
		if ! magick compare -metric SSIM -compose src \
			"http/$1.jpg" ".5-http/$1.jpg" ".5-http/$1-diff.jpg";
		then
			echo "ERROR: http/$1.jpg differs from .5-http/$1.jpg" \
				> ".5-http/$1.failed"
			echo "  accept: cp 'test/.5-http/$1.jpg' 'test/http/$1.jpg'" \
				>> ".5-http/$1.failed"
			exit 1
		fi
	fi

	if [ "$CHECKED" -ne 1 ]; then
		echo "http/$1 unknown test format" \
			> ".5-http/$1.failed"
		exit 1
	fi
	echo
}
export -f http_get_and_compare

if ! ls --zero http/* | sed -z 's/\.[^.]*//g' | sort -z | uniq -z \
	| parallel -0 --results .5-http-logs/{/.} \
	http_get_and_compare {/.} ;
then
	cat .5-http/*.failed
	fail
fi

kill $PID


success
