Implementing [Space Traders](https://spacetraders.io) API in Go.

Includes a cli to execute the API (and potentially play the game):

## Sample CLI

```
$ go run cli/cli.go
> help
Available commands:
<arguments> are required, [options] are optional.

    Help: Help [command]

  Account:
    Account: Account
    Claim: Claim <username> <path/to/file>
    Login: Login [path/to/file]
    Logout: Logout

  Loans:
    AvailableLoans: AvailableLoans
    MyLoans: MyLoans
    TakeLoan: TakeLoan <type>

  Ships:
    BuyShip: BuyShip <location> <type>
    ListShips: ListShips <system> [filter]
    MyShips: MyShips [filter]

  Locations:
    System: System [symbol]

  Goods and Cargo:
    Buy: Buy <shipID> <good> <quantity>
    Market: Market <location>


> help claim
Claim: Claim <username> <path/to/file>
Claims a username, saves token to specified file

> claim test15256 /tmp/test.readme

> account
test15256: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 14:38:10.477 +0000 UTC
> availableloans
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> takeLoan STARTUP
> account
test15256: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 14:38:10.477 +0000 UTC
> exit
```

## Implemented endpoints


* Game status - `/game/status`

* List all systems - `/game/systems`

* Available offers - `/locations/LOCATION/marketplace`

* Account details - `/my/account`

* List outstanding loans - `/my/loans`

* Take out loan - `/my/loans`

* Buy cargo - `/my/purchase-orders`

* Buy ship - `/my/ships`

* List my ship - `/my/ships`

* List ships for purchase - `/systems/LOCATION/ship-listing`

* Available loans - `/types/loans`

* Claim username - `/users/USERNAME/claim`

