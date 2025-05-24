#!/usr/bin/env python3

"""
Open two files, the first one has regexps, and compare them line by line.
"""
import re
import sys

refname = sys.argv[1]
regexps = open(sys.argv[1]).readlines()
fname = sys.argv[2]
lines = open(fname).readlines()
if len(regexps) != len(lines):
    print(f"{fname} has {len(lines)} lines, but {refname} has {len(regexps)} regexps")
    sys.exit(1)

for i, (regexp, line) in enumerate(zip(regexps, lines)):
    regexp = regexp.strip()
    line = line.strip()
    if not re.match(regexp, line):
        print(f"{fname}:{i + 1} does not match:")
        print(f"{fname}:{i + 1}   expected:  {regexp}")
        print(f"{fname}:{i + 1}   got:       {line}")
        sys.exit(1)

