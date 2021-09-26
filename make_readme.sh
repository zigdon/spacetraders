#!/bin/bash 

rm README.md /tmp/test.readme

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
takeLoan STARTUP
account
exit
"
echo "$CMDS" | go run cli/cli.go --echo 2>&1 | sed 's/> \(.*\)/> **\1**/' >> README.md

echo '```

## Implemented endpoints

' >> README.md

grep "// ##ENDPOINT" spacetraders.go | sed 's/.*ENDPOINT //' | sort | while read L ; do
  echo "* $L" >> README.md
  echo "" >> README.md
done

