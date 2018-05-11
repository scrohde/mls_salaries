package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type mlsData struct {
	Club         Club
	Name         string
	Pos          string
	BaseSalary   float64
	Compensation float64
}

// Clubs is a list of MLS clubs
type Clubs []Club

// Club is the abreviated and full name of an MLS club
type Club struct {
	Name     string
	FullName string
}

// Set sets the value of clubs
func (c *Clubs) Set(s string) error {
	names := strings.Split(s, ",")
loop:
	for _, name := range names {
		name = strings.TrimSpace(strings.ToUpper(name))
		for _, club := range allClubs {
			if club.Name == name || club.FullName == name {
				*c = append(*c, Club{club.Name, club.FullName})
				continue loop
			}
		}
		return fmt.Errorf("\ninvalid club: %s\nvalid clubs: %s", name, allClubs.String())
	}
	return nil
}

// String returns club names as a comma seperated list of abreviated names
func (c *Clubs) String() string {
	names := make([]string, len(*c), len(*c))
	for _, club := range *c {
		names = append(names, club.Name)
	}
	return strings.Join(names, ", ")
}

func (c *Clubs) hasVal(s string) bool {
	for _, val := range *c {
		if val.Name == s || val.FullName == s {
			return true
		}
	}
	return false
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

var allClubs = Clubs{
	{"NE", "New England Revolution"},
	{"ORL", "Orlando City SC"},
	{"SJ", "San Jose Earthquakes"},
	{"VAN", "Vancouver Whitecaps"},
	{"CLB", "Columbus Crew"},
	{"DC", "DC United"},
	{"MNUFC", "Minnesota United"},
	{"SEA", "Seattle Sounders FC"},
	{"CHI", "Chicago Fire"},
	{"COL", "Colorado Rapids"},
	{"DAL", "FC Dallas"},
	{"KC", "Sporting Kansas City"},
	{"LA", "LA Galaxy"},
	{"LAFC", "LAFC"},
	{"MTL", "Montreal Impact"},
	{"NYRB", "New York Red Bulls"},
	{"TOR", "Toronto FC"},
	{"ATL", "Atlanta United"},
	{"HOU", "Houston Dynamo"},
	{"NYCFC", "New York City FC"},
	{"PHI", "Philadelphia Union"},
	{"POR", "Portland Timbers"},
	{"RSL", "Real Salt Lake"},
}

// commaf returns v as a string with commas added
func commaf(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.Write([]byte{'-'})
		v = 0 - v
	}

	comma := []byte{','}

	parts := strings.Split(strconv.FormatFloat(v, 'f', 2, 64), ".")
	pos := 0
	if len(parts[0])%3 != 0 {
		pos += len(parts[0]) % 3
		buf.WriteString(parts[0][:pos])
		buf.Write(comma)
	}
	for ; pos < len(parts[0]); pos += 3 {
		buf.WriteString(parts[0][pos : pos+3])
		buf.Write(comma)
	}
	buf.Truncate(buf.Len() - 1)

	if len(parts) > 1 {
		buf.Write([]byte{'.'})
		buf.WriteString(parts[1])
	}
	return buf.String()
}

func main() {
	var all []mlsData
	var clubs Clubs

	flag.Var(&clubs, "clubs", "comma separated list of mls clubs")
	club := flag.Bool("sort", true, "sort by club")
	data := flag.String("data", "2018_05_01_data", "data file")
	flag.Parse()

	if len(clubs) == 0 {
		clubs = allClubs
	}

	f, err := os.Open(*data)
	if err != nil {
		panic(err)
	}

	const (
		FIRSTNAME = iota
		LASTNAME
		CLUB
		POSITION
		BASESALARY
		COMPENSATION
	)

	scanner := bufio.NewScanner(f)
	clubTotals := make(ClubTotals, 30)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), "\t")
		if len(tokens) < COMPENSATION+1 {
			continue
		}
		if !clubs.hasVal(tokens[CLUB]) {
			continue
		}

		data := mlsData{}
		for _, fullclub := range allClubs {
			if fullclub.Name == tokens[CLUB] || fullclub.FullName == tokens[CLUB] {
				data.Club.Name = fullclub.Name
				data.Club.FullName = fullclub.FullName
				break
			}
		}
		data.Pos = tokens[POSITION]
		data.Name = fmt.Sprintf("%s %s", tokens[FIRSTNAME], tokens[LASTNAME])
		data.BaseSalary, err = strconv.ParseFloat(strings.Replace(tokens[BASESALARY][1:], ",", "", -1), 32)
		data.Compensation, err = strconv.ParseFloat(strings.Replace(tokens[COMPENSATION][1:], ",", "", -1), 32)
		all = append(all, data)

		clubTotals[data.Club.Name] += data.Compensation
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Compensation > all[j].Compensation })
	if *club {
		sort.SliceStable(all, func(i, j int) bool { return all[i].Club.Name < all[j].Club.Name })
	}

	i := 1
	lastClub := all[0].Club
	for _, data := range all {
		if *club && data.Club != lastClub {
			i = 1
			lastClub = data.Club
			fmt.Println()
		}
		fmt.Printf("%-3d %-5s %-25s: %s\n", i, data.Club.Name, data.Name, commaf(data.Compensation))
		i++
	}

	fmt.Print("\n\n")
	for i, v := range clubTotals.Sort() {
		fmt.Printf("%-2d %-5s total: %s\n", i+1, v.Key, commaf(v.Value))
	}
}
