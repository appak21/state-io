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
		fmt.Fprintf(os.Stderr, "city-%d, %v\n", i, cities[i])
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
		myCities := make([]City, 0, cityN)
		for _, c := range cities {
			if c.ownerID == playerID {
				myCities = append(myCities, c)
			}
		}
		myCityN := len(myCities)
		// sort my cities by max units
		sort.SliceStable(myCities, func(i, j int) bool {
			return myCities[i].units > myCities[j].units
		})

		hisCities := make([]City, 0, cityN)
		for _, c := range cities {
			if c.ownerID != playerID && c.ownerID != NEUTRAL {
				hisCities = append(hisCities, c)
			}
		}
		// sort his cities by max units
		sort.SliceStable(hisCities, func(i, j int) bool {
			return hisCities[i].units > hisCities[j].units
		})

		neutralCities := make([]City, 0, cityN)
		for _, c := range cities {
			if c.ownerID == NEUTRAL {
				neutralCities = append(neutralCities, c)
			}
		}
		//-------------------MAKING DECISION-------------------------
		maxPrize := MININT
		src, dest := coords{}, coords{}
		for _, mc := range myCities {
			for _, c := range cities {
				if mc == c {
					continue
				}
				prize := conquer(mc, c, movements)
				if prize > 0 && !isSafe(mc, c, hisCities) {
					prize = 0
				}
				if maxPrize <= prize {
					maxPrize = prize
					dest = c.coords
					src = mc.coords
				}
			}
			if mc.units == MAXUNITS-myCityN && maxPrize <= 0 {
				sort.SliceStable(hisCities, func(i, j int) bool {
					d1 := distance(mc.coords, hisCities[i].coords)
					d2 := distance(mc.coords, hisCities[j].coords)
					return d1 < d2
				})
				for _, city := range hisCities {
					if city.units <= mc.units {
						src = mc.coords
						dest = city.coords
						break
					}
				}
				maxPrize = 1
			}
		}

		if maxPrize <= 0 {
			largestCity := myCities[0]
			if len(myCities)-len(hisCities) > 1 {
				sort.SliceStable(neutralCities, func(i, j int) bool {
					d1 := distance(largestCity.coords, neutralCities[i].coords)
					d2 := distance(largestCity.coords, neutralCities[j].coords)
					return d1 < d2
				})
				for _, city := range neutralCities {
					if city.units < largestCity.units {
						src = largestCity.coords
						dest = city.coords
						break
					}
				}
			} else {
				dest = coords{}
			}
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
		if myCost > hisCost {
			return false
		}
	}
	return true
}

func conquer(src, dest City, movements []Movement) int {
	dist := distance(src.coords, dest.coords) + 1
	cityUnits1 := getCityUnits(src, dest, movements)
	if cityUnits1 > 0 { // CITY IS MINE / NEUTRAL or WILL BE MINE or MY / NEUTRAL CITY IS NOT UNDER ATTACK
		return 0
	}
	// CITY IS HIS or WILL BE HIS or HIS CITY IS NOT UNDER ATTACK
	m := Movement{toCity: dest.name, attackerID: src.ownerID, leftTicks: dist, units: src.units}
	cityUnits2 := getCityUnits(src, dest, Insert(movements, m))
	if cityUnits2 > 0 {
		cityUnits2 += OCCUPIEDCITY - dist
	}
	return cityUnits2
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

// returns the most optimal neutral city for myCity, only when opponent has only 1 city
// that should be close to me, far from my opponent
func getNeutral(myCity, hisCity City, neutralCities []City) coords {
	if len(neutralCities) == 0 {
		return coords{}
	}
	max := MININT
	theCity := coords{}
	for _, nc := range neutralCities {
		myDist := distance(myCity.coords, nc.coords)
		hisDist := distance(hisCity.coords, nc.coords)
		if hisDist-myDist > max {
			max = hisDist - myDist
			theCity = nc.coords
		}
	}
	return theCity
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
