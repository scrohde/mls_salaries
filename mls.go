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

// Player is an MLS player
type Player struct {
	Club         Club
	Name         string
	Pos          string
	BaseSalary   float64
	Compensation float64
}

// Players is a list of MLS Players
type Players []Player

// Set sets the value of Players from a comma separated list
func (p *Players) Set(s string) error {
	names := strings.Split(s, ",")
	for _, name := range names {
		*p = append(*p, Player{Name: strings.ToLower(strings.TrimSpace(name))})
	}
	return nil
}

func (p *Players) String() string {
	names := make([]string, len(*p), len(*p))
	for _, player := range *p {
		names = append(names, player.Name)
	}
	return strings.Join(names, ", ")
}

func (p *Players) HasVal(s string) bool {
	for _, player := range *p {
		if strings.Contains(strings.ToLower(s), player.Name) {
			return true
		}
	}
	return false
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

type DPs []string

var allDPs = DPs{
	"Fanendo Adi",
	"Romain Alessandrini",
	"Miguel Almiron",
	"Jozy Altidore",
	"Paul Arriola",
	"Ezequiel Barco",
	"Alejandro Bedoya",
	"Sebastian Blanco",
	"Michael Bradley",
	"Josue Colman",
	"Cristian Colman",
	"Yohan Croizet",
	"Clint Dempsey",
	"Claude Dielna",
	"Borek Dockal",
	"Giovani dos Santos",
	"Jonathan dos Santos",
	"Dom Dwyer",
	"Alberth Elis",
	"Shkelzen Gashi",
	"Sebastian Giovinco",
	"Carlos Gruezo",
	"Felipe Gutierrez",
	"Federico Higuain",
	"Andre Horta",
	"Tim Howard",
	"Sacha Kljestan",
	"Nicolas Lodeiro",
	"Josef Martinez",
	"Tomas Martinez",
	"Jesus Medina",
	"Lucas Melano",
	"Maxi Moralez",
	"Santiago Mosquera",
	"Nemanja Nikolic",
	"Ignacio Piatti",
	"Valeri \"Vako\" Qazaishvili",
	"Darwin Quintero",
	"Angelo Rodriguez",
	"Wayne Rooney",
	"Alejandro 'Kaku' Romero",
	"Diego Rossi",
	"Raul Ruidiaz",
	"Johnny Russell",
	"Albert Rusnak",
	"Pedro Santos",
	"Jefferson Savarino",
	"Bastian Schweinsteiger",
	"Brek Shea",
	"Saphir Taider",
	"Erick Torres",
	"Milton Valenzuela",
	"Diego Valeri",
	"Carlos Vela",
	"David Villa",
	"Kendall Waston",
	"Chris Wondolowski",
	"Bradley Wright-Phillips",
}

func (d DPs) hasVal(name string) bool {
	for _, val := range d {
		if val == name {
			return true
		}
	}
	return false
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
	var (
		all     Players
		clubs   Clubs
		players Players
	)

	flag.Var(&clubs, "clubs", "comma separated list of mls clubs")
	flag.Var(&players, "players", "comma separated list of mls players")
	club := flag.Bool("sort", true, "sort by club")
	dps := flag.Bool("dp", false, "only show DP players")
	data := flag.String("data", "2018_09_15_data", "data file")
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

		player := Player{}
		for _, fullclub := range allClubs {
			if fullclub.Name == tokens[CLUB] || fullclub.FullName == tokens[CLUB] {
				player.Club.Name = fullclub.Name
				player.Club.FullName = fullclub.FullName
				break
			}
		}
		player.Pos = tokens[POSITION]
		player.Name = fmt.Sprintf("%s %s", tokens[FIRSTNAME], tokens[LASTNAME])
		player.BaseSalary, err = strconv.ParseFloat(strings.Replace(tokens[BASESALARY][1:], ",", "", -1), 32)
		player.Compensation, err = strconv.ParseFloat(strings.Replace(tokens[COMPENSATION][1:], ",", "", -1), 32)

		if players != nil && !players.HasVal(player.Name) {
			continue
		}

		if *dps && !allDPs.hasVal(player.Name) {
			continue
		}

		all = append(all, player)
		clubTotals[player.Club.Name] += player.Compensation
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
