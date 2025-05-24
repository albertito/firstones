#!/usr/bin/env python3

"""Extract the IPA symbols from the dictionary files.

This script reads all .txt files in the current directory, and then extracts
all the IPA symbols used in them.

It is used to manually update the ipa.go/ipaToGlyphs map.
"""

from collections import defaultdict
import re
import os

dicts = []
s = ""
for f in os.listdir("."):
    if f.endswith(".txt"):
        dicts.append(f.strip(".txt"))
        s += open(f).read()

chars = defaultdict(int)
for p in re.finditer(r"/(\w+)/", s):
    for c in p.group(1):
        chars[c] += 1

print("Dicts:", ", ".join(dicts))
print("Frequency:")
total = sum(chars.values())
for c, count in sorted(chars.items(), key=lambda x: x[1], reverse=True):
    pct = count / total * 100
    print(f"  {c}: {count} ({pct:.2f}%)")

print("Count:", len(chars))
print("Chars:", "".join(sorted(chars)))

# Symbols we already know about. The ones present in ipa.go/ipaToGlyphs map.
known = "abdefghijklmnoprstuvwxzæðŋɑɔəɛɝɡɣɪɫɲɹɾʃʊʎʒʝˈˌβθ"

# Print the unknown symbols in an easy to copy-paste way.
unknown = sorted(set(chars.keys()) - set(known))
print(len(unknown), "unknown:", "".join(unknown))
for c in unknown:
    print(f"  '{c}': \"\",")
