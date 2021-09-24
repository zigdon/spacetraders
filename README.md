Implementing [Space Traders](https://spacetraders.io) API in Go.

Includes a cli to execute the API (and potentially play the game):

```
$ go run cli/cli.go
2021/09/24 09:20:21 Status: spacetraders is currently online and available to play
> help
Available commands: account, availableLoans, claim, help, listShips, login, logout, myLoans, system, takeLoan
> help claim
claim username path/to/file:
Claims a username, saves token to specified file
> claim test8880 /tmp/test8880

> account
test8880: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-24 16:20:48.69 +0000 UTC
> availableLoans
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> takeLoan STARTUP
Loan taken, id=cktykmqzs81868615s6nt83awbp, due: 2021-09-26 16:21:13.383 +0000 UTC
> account
test8880: Credits: 200000, Ships: 0, Structures: 0, Joined: 2021-09-24 16:20:48.69 +0000 UTC
> exit
```
