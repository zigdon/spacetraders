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

  Flight Plans:
    CreateFlightPlan: CreateFlightPlan <shipID> <destination>
    ShowFlightPlan: ShowFlightPlan <flightPlanID>

  Locations:
    System: System [system]

  Goods and Cargo:
    Buy: Buy <shipID> <good> <quantity>
    Market: Market <location>


> help claim
Claim: Claim <username> <path/to/file>
Claims a username, saves token to specified file

> claim test31634 /tmp/test.readme

> account
test31634: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 12:58:23.286 -0700 PDT
> availableloans
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> takeloan STARTUP
Loan taken, ln-1 (cku1n9qhb7403515s6meecvxkp), due: 2021-09-28 19:58:23.565 +0000 UTC
> listships OE MK-I
JW-MK-I: Jackshaw MK-I
speed: 1, cargo: 50, weapons: 5, plating: 5
  OE-PM-TR: 21125
GR-MK-I: Gravager MK-I
speed: 1, cargo: 100, weapons: 5, plating: 10
  OE-PM-TR: 42650
EM-MK-I: Electrum MK-I
speed: 2, cargo: 50, weapons: 10, plating: 5
  OE-PM-TR: 37750
HM-MK-I: Hermes MK-I
speed: 3, cargo: 50, weapons: 5, plating: 20
  OE-PM-TR: 57525
TD-MK-I: Tiddalik MK-I
speed: 2, cargo: 3000, weapons: 5, plating: 10
  OE-UC-AD: 473600

> buyship OE-PM-TR JW-MK-I
New ship ID: s-1 (cku1n9qzp7418315s6vm2p7q1i)
> myships
s-1: Jackshaw MK-I (JW-MK-I)
ID: cku1n9qzp7418315s6vm2p7q1i
Speed: 1, Max cargo: 50, Available space: 50, Weapons: 5, Plating: 5
At OE-PM-TR (14, 18)

> buy s-1 FUEL 20
Bought 20 of FUEL for 60

> myships s-1
s-1: Jackshaw MK-I (JW-MK-I)
ID: cku1n9qzp7418315s6vm2p7q1i
Speed: 1, Max cargo: 50, Available space: 30, Weapons: 5, Plating: 5
At OE-PM-TR (14, 18)
Cargo:
  20 of FUEL (20)

> exit
```

### Caching

The cli uses a cache to do argument checking for commands, e.g. `ListShips`
will only accept known systems as an argument, while `Market` only takes
locations where you have ships.

This behaviour can be disabled by passing `--nocache` to the cli.

## Implemented endpoints


* Game status - `/game/status`

* List all systems - `/game/systems`

* Available offers - `/locations/LOCATION/marketplace`

* Account details - `/my/account`

* Create flight plan - `/my/flight-plans`

* Show flight plans - `/my/flight-plans/FLIGHTID`

* List outstanding loans - `/my/loans`

* Take out loan - `/my/loans`

* Buy cargo - `/my/purchase-orders`

* Buy ship - `/my/ships`

* List my ship - `/my/ships`

* List ships for purchase - `/systems/LOCATION/ship-listing`

* Available loans - `/types/loans`

* Claim username - `/users/USERNAME/claim`

