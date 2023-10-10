package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

const (
	NEUTRAL      = "Neutral"
	MAXUNITS     = 50
	OCCUPIEDCITY = 20
	TICK         = 1.74
	MININT       = -1111
)

type coords struct {
	x, y int
}

type City struct {
	name    string
	ownerID string
	units   int
	coords
}

type Movement struct {
	fromCity    string
	toCity      string
	attackerID  string
	cityOwnerID string
	leftTicks   int
	units       int
}

// returns cities
func readCities(cityN int) []City {
	cities := make([]City, cityN)
	for i := 0; i < cityN; i++ {
		data := ""
		fmt.Scan(&cities[i].ownerID, &cities[i].units, &cities[i].name, &cities[i].x, &cities[i].y, &data, &data)
		// fmt.Fprintf(os.Stderr, "city-%d, %v\n", i, cities[i])
	}
	return cities
}

// returns movements
func readMovements() []Movement {
	var movementN int
	fmt.Scan(&movementN)
	fmt.Fprintf(os.Stderr, "movementN=%d\n", movementN)
	movements := make([]Movement, movementN)
	for i := 0; i < movementN; i++ {
		fmt.Scan(&movements[i].fromCity, &movements[i].toCity, &movements[i].attackerID, &movements[i].cityOwnerID, &movements[i].leftTicks, &movements[i].units)
		fmt.Fprintln(os.Stderr, movements[i].fromCity, movements[i].toCity, movements[i].attackerID, movements[i].cityOwnerID, movements[i].leftTicks, "-tick(s) left", movements[i].units)
	}
	return movements
}

func move(src, dest coords) {
	if dest.x == 0 {
		fmt.Fprintln(os.Stderr, "No move")
		fmt.Println(" ")
		return
	}
	fromX, fromY := strconv.Itoa(src.x), strconv.Itoa(src.y)
	toX, toY := strconv.Itoa(dest.x), strconv.Itoa(dest.y)
	fmt.Fprintln(os.Stderr, fromX+" "+fromY+" "+toX+" "+toY)
	fmt.Println(fromX + " " + fromY + " " + toX + " " + toY)
}

func main() {
	for true {
		var cityN, tick int
		var playerID string
		fmt.Scan(&cityN, &playerID, &tick)
		cities := readCities(cityN)

		movements := readMovements()
		//-------------------PREPARING DATA-------------------------
		sort.SliceStable(movements, func(i, j int) bool {
			return movements[i].leftTicks < movements[j].leftTicks
		})

		myCities := make([]City, 0, cityN)
		hisCities := make([]City, 0, cityN)
		neutralCities := make([]City, 0, cityN)

		for _, c := range cities {
			if c.ownerID == playerID {
				myCities = append(myCities, c)
			} else if c.ownerID == NEUTRAL {
				neutralCities = append(neutralCities, c)
			} else {
				hisCities = append(hisCities, c)
			}
		}
		// sort my cities by max units
		sort.SliceStable(myCities, func(i, j int) bool {
			return myCities[i].units > myCities[j].units
		})
		// sort his cities by max units
		sort.SliceStable(hisCities, func(i, j int) bool {
			return hisCities[i].units > hisCities[j].units
		})

		//-------------------MAKING DECISION-------------------------
		minDist := 1000
		maxPrize, time := MININT, 0
		src, dest := coords{}, coords{}
		for _, mc := range myCities {

			for _, c := range cities {
				if mc == c {
					continue
				}
				prize, mytime := conquer(mc, c, movements)
				if prize > 0 && !isSafe(mc, c, hisCities) {
					prize = 0
				}
				if maxPrize <= prize {
					time = mytime
					maxPrize = prize
					dest = c.coords
					src = mc.coords
				}
			}

			if ok, t := isSafeToLeave1(mc, movements); !ok {
				if time > t {
					maxPrize = 0
					continue
				}
			}

			if maxPrize <= 0 {
				for i := 0; i < len(hisCities); i++ {
					temp := distance(mc.coords, hisCities[i].coords)
					if minDist > temp {
						minDist = temp
						src = mc.coords
						dest = hisCities[i].coords
					}
				}
			}
		}
		if maxPrize <= 0 && len(hisCities) == 0 {
			dest = coords{}
		}
		move(src, dest)
	}
}

func isSafe(src, dest City, hisCities []City) bool {
	d := distance(src.coords, dest.coords)
	myCost := 2*d - 1
	for _, c := range hisCities {
		if c == dest {
			continue
		}
		d = distance(c.coords, src.coords) + 1
		hisCost := 2 * d
		if myCost > hisCost && c.units > d {
			return false
		}
	}
	return true
}

func isSafeToLeave1(src City, movements []Movement) (bool, int) {
	cityUnits := 0
	lastLeftTicks, sign := 0, 1
	for _, m := range movements {
		if src.name == m.toCity {
			cost := m.leftTicks - lastLeftTicks
			if cityUnits == MAXUNITS || cityUnits == -MAXUNITS {
				cost = 0
			}
			if m.attackerID == src.ownerID { // PROTECTING
				cityUnits = m.units + cityUnits + sign*cost
			} else { // HIM
				cityUnits = -(m.units - cityUnits - sign*cost)
			}
		}
		if cityUnits < 0 {
			sign = -1
		} else if cityUnits > 0 {
			sign = 1
		}
		lastLeftTicks = m.leftTicks
	}
	if cityUnits >= 0 {
		return true, -1
	}
	return !isSafeToLeave2(src, movements), lastLeftTicks
}

func isSafeToLeave2(src City, movements []Movement) bool {
	cityUnits := src.units
	lastLeftTicks, sign := 0, 1
	for _, m := range movements {
		if src.name == m.toCity {
			cost := m.leftTicks - lastLeftTicks
			if cityUnits == MAXUNITS || cityUnits == -MAXUNITS {
				cost = 0
			}
			if m.attackerID == src.ownerID { // ME
				cityUnits = m.units + cityUnits + sign*cost
			} else { // HIM
				cityUnits = -(m.units - cityUnits - sign*cost)
			}
		}
		if cityUnits < 0 {
			sign = -1
		} else if cityUnits > 0 {
			sign = 1
		}
		lastLeftTicks = m.leftTicks
	}
	return cityUnits >= 0
}

func conquer(src, dest City, movements []Movement) (int, int) {
	dist := distance(src.coords, dest.coords) + 1
	// fmt.Fprintln(os.Stderr, " DISTANCE ", dist-1, " from ", src.coords, " to ", dest.coords)
	cityUnits1 := getCityUnits(src, dest, movements)
	if cityUnits1 > 0 { // CITY IS MINE / NEUTRAL or WILL BE MINE or MY / NEUTRAL CITY IS NOT UNDER ATTACK
		return 0, -1
	}
	// CITY IS HIS or WILL BE HIS or HIS CITY IS NOT UNDER ATTACK
	m := Movement{toCity: dest.name, attackerID: src.ownerID, leftTicks: dist, units: src.units}
	cityUnits2 := getCityUnits(src, dest, Insert(movements, m))
	if cityUnits2 > 0 {
		cityUnits2 = OCCUPIEDCITY - dist
	}
	return cityUnits2, dist
}

// cityUnits is a positive num when city is mine, and negative for opponent's city
// the func is not for Neutral city
func getCityUnits(src, dest City, movements []Movement) int {
	isNeutral := true
	lastLeftTicks := 0
	sign := 1
	if dest.ownerID != src.ownerID && dest.ownerID != NEUTRAL {
		sign = -1
	}
	cityUnits := sign * dest.units
	for _, m := range movements {
		if dest.name == m.toCity { // CITY IS UNDER ATTACK
			cost := m.leftTicks - lastLeftTicks
			if dest.ownerID == NEUTRAL && isNeutral {
				tmp := m.units - cityUnits       // hope all neutral city units is 10
				if m.attackerID == src.ownerID { // if ATTACKER is me
					cityUnits = tmp
				} else { // attacker is him
					cityUnits = -tmp
					sign = -1
				}
				isNeutral = false
				continue
			}
			if cityUnits == MAXUNITS || cityUnits == -MAXUNITS {
				cost = 0
			}
			if m.attackerID == src.ownerID { // if ATTACKER is me
				cityUnits = m.units + cityUnits + sign*cost
			} else { // attacker is him
				cityUnits = -(m.units - cityUnits - sign*cost)
			}
		}

		if cityUnits < 0 {
			sign = -1
		} else if cityUnits > 0 {
			sign = 1
		}
		lastLeftTicks = m.leftTicks
	}
	return cityUnits
}

// calculates the distance between src and dest coords
func distance(src, dest coords) int {
	x1, y1 := src.x/100, src.y/100
	x2, y2 := dest.x/100, dest.y/100
	res := math.Sqrt(math.Pow(float64(x2)-float64(x1), 2) + math.Pow(float64(y2)-float64(y1), 2)*1.0)
	return int(math.Round(res * TICK))
}

func Insert(mm []Movement, m Movement) []Movement {
	mm = append(mm, Movement{})
	i := BinarySearch(mm, m.leftTicks)
	copy(mm[i+1:], mm[i:])
	mm[i] = m
	return mm
}

func BinarySearch(m []Movement, target int) int {
	i, j := 0, len(m)-1
	for i < j {
		h := (i + j) >> 1
		if m[h].leftTicks < target {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
