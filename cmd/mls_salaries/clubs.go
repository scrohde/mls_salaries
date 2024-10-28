package main

import (
	"fmt"
	"sort"
	"strings"
)

// Clubs is a map of MLS club names to abbreviated names
type Clubs map[string]string

var allClubs = Clubs{
	"Pool":                   "MLS",
	"MLS Pool":               "MLS",
	"Major League Soccer":    "MLS",
	"Retired":                "MLS",
	"New England Revolution": "NE",
	"Orlando City SC":        "ORL",
	"San Jose Earthquakes":   "SJ",
	"Vancouver Whitecaps":    "VAN",
	"Columbus Crew":          "CLB",
	"DC United":              "DC",
	"Minnesota United":       "MNUFC",
	"Seattle Sounders FC":    "SEA",
	"Chicago Fire":           "CHI",
	"Colorado Rapids":        "COL",
	"FC Dallas":              "DAL",
	"Sporting Kansas City":   "KC",
	"LA Galaxy":              "LA",
	"LAFC":                   "LAFC",
	"CF Montreal":            "MTL",
	"Montreal":               "MTL",
	"Montreal Impact":        "MTL",
	"New York Red Bulls":     "NYRB",
	"Toronto FC":             "TOR",
	"Atlanta United":         "ATL",
	"Houston Dynamo":         "HOU",
	"New York City FC":       "NYCFC",
	"Philadelphia Union":     "PHI",
	"Portland Timbers":       "POR",
	"Real Salt Lake":         "RSL",
	"FC Cincinnati":          "CIN",
	"NY":                     "NYRB",
	"Chivas USA":             "CHV",
	"Nashville SC":           "NSC",
	"Inter Miami":            "MIA",
	"Austin FC":              "AFC",
	"Charlotte FC":           "CLT",
	"St. Louis SC":           "STL",
	"St. Louis City SC":      "STL",
	"San Diego FC":           "SDFC",
}

// Set sets the value of clubs
func (c *Clubs) Set(s string) error {
	*c = make(Clubs)
	for _, name := range strings.Split(s, ",") {
		name = strings.TrimSpace(strings.ToUpper(name))
		if key, ok := allClubs.getKey(name); ok {
			(*c)[key] = name
		} else {
			return fmt.Errorf("valid clubs: %s", allClubs.String())
		}
	}
	return nil
}

func (c *Clubs) getKey(val string) (string, bool) {
	for key, value := range *c {
		if val == value {
			return key, true
		}
	}
	return "", false
}

// HasVal returns true if s is the full or abbreviated name of a club
func (c *Clubs) HasVal(val string) bool {
	if _, ok := (*c)[val]; ok {
		return true
	}
	_, ok := (*c).getKey(val)
	return ok
}

// Abv returns the abbreviated name of a club
func (c *Clubs) Abv(fullName string) (abvName string) {
	if abv, ok := (*c)[fullName]; ok {
		return abv
	}
	if _, ok := (*c).getKey(fullName); ok {
		return fullName
	}
	return ""
}

// String returns club names as a comma separated list of abbreviated names
func (c *Clubs) String() string {
	var names []string
	for _, val := range *c {
		names = append(names, val)
	}
	return strings.Join(names, ", ")
}

// ClubTotals maps club names to total compensation
type ClubTotals map[string]float64

// KeyValue holds a key/value pair
type KeyValue struct {
	Key   string
	Value float64
}

// Sort returns a sorted slice of ClubTotals key/value pairs
func (ct *ClubTotals) Sort() []KeyValue {
	p := make([]KeyValue, len(*ct))
	i := 0
	for k, v := range *ct {
		p[i] = KeyValue{k, v}
		i++
	}
	sort.Slice(p, func(i, j int) bool { return p[i].Value > p[j].Value })
	return p
}
