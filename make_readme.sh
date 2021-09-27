#!/bin/bash 

set -e

rm -f README.md /tmp/test.readme

cat <<'README' > README.md
Implementing [Space Traders](https://spacetraders.io) API in Go.

Includes a cli to execute the API (and potentially play the game):

## Sample CLI

```
$ go run bin/spacetraders.go
README

CMDS="
help
help claim
claim test$RANDOM /tmp/test.readme
account
availableloans
takeloan STARTUP
listships OE MK-I
buyship OE-PM-TR JW-MK-I
myships
buy s-1 FUEL 20
buy s-1 METALS 25
myships s-1
locations oe
createflightplan s-1 OE-PM
showflightplan f-1
wait f-1
sell s-1 METALS 25
exit
"
echo "$CMDS" | go run bin/spacetraders.go --errors_fatal --debug --echo >> README.md

echo '```

### Short IDs

Most cases where an object ID is required (e.g. `cku26s3jz800715s6siwejax8`), a
short ID is generated that can be used instead (e.g. `s-2`, `f-1` for the 2nd
ship and the first flight plan, respectively).

In addition, a prefix is sufficient for any ID, as long as it is unique for
that object type. If you have two ships, with the following IDs:

* `cku26s3jz800715s6siwejax8`
* `cku26s4a7824215s6iyyhozhp`

They could be referenced as `cku26s3` and `cku26s4`.

### Caching

The cli uses a cache to do argument checking for commands, e.g. `ListShips`
will only accept known systems as an argument, while `Market` only takes
locations where you have ships.

This behaviour can be disabled by passing `--nocache` to the cli.

## Implemented endpoints

' >> README.md

grep "// ##ENDPOINT" spacetraders.go | sed 's/.*ENDPOINT //' | sort -t\- -k2 | while read L ; do
  echo "* $L" >> README.md
  echo "" >> README.md
done

echo ===================================
cat README.md
