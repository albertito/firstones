#!/bin/bash

if [ "$V" == "1" ]; then
	set -v
fi

TESTDIR=$( realpath "$(dirname "${BASH_SOURCE[0]}")" )
export TESTDIR
cd "${TESTDIR}" || exit 1

# Set traps to kill our subprocesses when we exit (for any reason).
trap ":" TERM      # Avoid the EXIT handler from killing bash.
trap "exit 2" INT  # Ctrl-C, make sure we fail in that case.
trap "kill 0" EXIT # Kill children on exit.

export GOCOVERDIR

# Wait until there's something listening on the given port.
function wait_until_ready() {
	PORT=$1

	while ! bash -c "true < /dev/tcp/localhost/$PORT" 2>/dev/null ; do
		sleep 0.01
	done
}

# Wait until grep returns true, or fail on timeout.
function wait_grep() {
	for i in 0.01 0.02 0.05 0.1 0.2; do
		if grep "$@"; then
			return 0
		fi
		sleep $i
	done
	return 1
}

function build() {
	(
		cd ..
		if [ "$GOCOVERDIR" != "" ]; then
			go build -cover -covermode=count
		else
			go build
		fi
	)
}

function success() {
	echo success
}

function fail() {
	echo "FAILED"
	exit 1
}

function fo() {
	../firstones "$@"
}
export -f fo

# Run in the background (sets $PID to its process id).
function fo_bg() {
	# Duplicate fo() because if we put the function in the background,
	# the pid will be of bash, not the subprocess.
	../firstones "$@" > .firstones.log 2>&1 &
	PID=$!
	export PID
}
