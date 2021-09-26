Implementing [Space Traders](https://spacetraders.io) API in Go.

Includes a cli to execute the API (and potentially play the game):

## Sample CLI

```
$ go run cli/cli.go
2021/09/26 07:32:26 Status: spacetraders is currently online and available to play
> **help**
Available commands:
<arguments> **are required, [options] are optional.**

    Help: Help [command]

  Account:
    Account: Account
    Claim: Claim <username> **<path/to/file>**
    Login: Login [path/to/file]
    Logout: Logout

  Loans:
    AvailableLoans: AvailableLoans
    MyLoans: MyLoans
    TakeLoan: TakeLoan <type>

  Ships:
    BuyShip: BuyShip <location> **<type>**
    ListShips: ListShips <system> **[filter]**
    MyShips: MyShips [filter]

  Locations:
    System: System [symbol]

  Goods and Cargo:
    Buy: Buy <shipID> **<good> <quantity>**
    Market: Market <location>


> **help claim**
Claim: Claim <username> **<path/to/file>**
Claims a username, saves token to specified file

> **claim test13686 /tmp/test.readme**
2021/09/26 07:32:26 Got token "e899c1de-cb69-4f07-a450-d8b5cf36701b" for "test13686"

> **account**
test13686: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 14:32:26.436 +0000 UTC
> **availableloans**
2021/09/26 07:32:26 Cashing 1 "loans" items for 1m0s
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> **takeLoan STARTUP**
2021/09/26 07:32:26 Unknown command "takeloan". Try 'help'.
> **account**
test13686: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 14:32:26.436 +0000 UTC
> **exit**
```

## Implemented endpoints


* Account details - `/my/account`

* Available loans - `/types/loans`

* Available offers - `/locations/LOCATION/marketplace`

* Buy cargo - `/my/purchase-orders`

* Buy ship - `/my/ships`

* Claim username - `/users/USERNAME/claim`

* Game status - `/game/status`

* List all systems - `/game/systems`

* List my ship - `/my/ships`

* List outstanding loans - `/my/loans`

* List ships for purchase - `/systems/LOCATION/ship-listing`

* Take out loan - `/my/loans`

