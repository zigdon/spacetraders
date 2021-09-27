#!/bin/bash 

set -e

rm -f README.md /tmp/test.readme

cat <<'README' > README.md
Implementing [Space Traders](https://spacetraders.io) API in Go.

Includes a cli to execute the API (and potentially play the game):

## Sample CLI

```
$ go run cli/cli.go
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
echo "$CMDS" | go run cli/cli.go --errors_fatal --debug --echo >> README.md

echo '```

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
