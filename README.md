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
    Wait: Wait <flightPlanID>

  Locations:
    Locations: Locations <system> [type]
    System: System [system]

  Goods and Cargo:
    Buy: Buy <shipID> <good> <quantity>
    Market: Market <location>
    Sell: Sell <shipID> <good> <quantity>


> help claim
Claim: Claim <username> <path/to/file>
Claims a username, saves token to specified file

> claim test25356 /tmp/test.readme

> account
test25356: Credits: 0, Ships: 0, Structures: 0, Joined: 2021-09-26 16:08:21.983 -0700 PDT
> availableloans
amt: 200000, needs collateral: false, rate: 40, term (days): 2, type: STARTUP
> takeloan STARTUP
Loan taken, ln-1 (cku1u21tk66101215s6w867952k), due: 2021-09-28 23:08:22.326 +0000 UTC
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
New ship ID: s-1 (cku1u22d166116115s6dpqoly18)
> myships
s-1: Jackshaw MK-I (JW-MK-I)
ID: cku1u22d166116115s6dpqoly18
Speed: 1, Max cargo: 50, Available space: 50, Weapons: 5, Plating: 5
At OE-PM-TR (14, 18)

> buy s-1 FUEL 20
Bought 20 of FUEL for 60

> buy s-1 METALS 25
Bought 25 of METALS for 100

> myships s-1
s-1: Jackshaw MK-I (JW-MK-I)
ID: cku1u22d166116115s6dpqoly18
Speed: 1, Max cargo: 50, Available space: 5, Weapons: 5, Plating: 5
At OE-PM-TR (14, 18)
Cargo:
  20 of FUEL (20)
  25 of METALS (25)

> locations oe
Using "OE" for "oe"
10 locations in "OE":
  OE-PM: Prime
    Type: PLANET  (13, 16)
    Traits: [METAL_ORES SOME_ARABLE_LAND]
  OE-PM-TR: Tritus
    Type: MOON  (14, 18)
    Traits: [METAL_ORES]
  OE-CR: Carth
    Type: PLANET  (10, 11)
    Traits: [METAL_ORES ARABLE_LAND RARE_METAL_ORES]
  OE-KO: Koria
    Type: PLANET  (-33, -36)
    Traits: [SOME_METAL_ORES SOME_NATURAL_CHEMICALS]
  OE-UC: Ucarro
    Type: PLANET  (74, -15)
    Traits: [SOME_METAL_ORES NATURAL_CHEMICALS]
  OE-UC-AD: Ado
    Type: MOON  (76, -14)
    Traits: [TECHNOLOGICAL_RUINS]
  OE-UC-OB: Obo
    Type: MOON  (77, -16)
    Traits: [NATURAL_CHEMICALS]
  OE-NY: Nyon
    Type: ASTEROID  (-58, 24)
    Allows construction.
  OE-BO: Bo
    Type: GAS_GIANT  (-60, -58)
    Allows construction.
    Traits: [SOME_HELIUM_3]
  OE-W-XV: Wormhole
    Type: WORMHOLE  (87, 55)
    Extensive research has revealed a partially functioning warp gate harnessing the power of an unstable but traversable wormhole.
    The scientific community has determined a means of stabilizing the ancient structure.
    Enter at your own risk.
    GET https://api.spacetraders.io/locations/OE-W-XV/structures
    POST https://api.spacetraders.io/structures/:structureId/deposit shipId=:shipId good=:goodSymbol quantity=:quantity
    POST https://api.spacetraders.io/my/warp-jumps shipId=:shipId

> createflightplan s-1 OE-PM
Created flight plan: f-1:  OE-PM-TR->OE-PM, ETA: 35s

> showflightplan f-1
f-1: cku1u22d166116115s6dpqoly18 OE-PM-TR->OE-PM
  ID: cku1u239u66156615s6s11xt35l
  Arrives at: 2021-09-26 23:09:00.206 +0000 UTC, ETA: 35.800345133s
  Fuel consumed: 1, remaining: 19
  Distance: 2, Terminated: 0001-01-01 00:00:00 +0000 UTC

> wait f-1
Waiting 34s for f-1 (cku1u239u66156615s6s11xt35l) to arrive... Arrived!

> sell s-1 METALS 25
Sold 25 of METALS from 1000

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

* Sell cargo - `/my/sell-orders`

* Buy ship - `/my/ships`

* List my ship - `/my/ships`

* List ships for purchase - `/systems/LOCATION/ship-listing`

* List locations in a system - `/systems/SYSTEM/locations`

* Available loans - `/types/loans`

* Claim username - `/users/USERNAME/claim`

