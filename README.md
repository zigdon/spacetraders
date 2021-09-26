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

> claim test26384 /tmp/test.readme

> account
test26384: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 16:08:07.166 +0000 UTC
> availableloans
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> takeloan STARTUP
Loan taken, id=cku1f1lxo105522515s6xvvutqdc, due: 2021-09-28 16:08:07.499 +0000 UTC
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
New ship ID: cku1f1mf1105532315s6mcksfb0u
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

