package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// Debug controls if debug info is printed
var Debug = false

//Printf prints if Debug is false
func Printf(format string, a ...interface{}) {
	if !Debug {
		fmt.Printf(format, a...)
	}
}

//Dprintln prints if Debug is true
func Dprintln(a ...interface{}) {
	if Debug {
		fmt.Println(a...)
	}
}

// Player is an MLS player
type Player struct {
	Club         string
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
		*p = append(*p, Player{Name: strings.TrimSpace(name)})
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

// HasVal return true if any players name contains s
func (p *Players) HasVal(s string) bool {
	for _, player := range *p {
		if strings.Contains(strings.ToLower(s), strings.ToLower(player.Name)) {
			return true
		}
	}
	return false
}

// Clubs is a map of MLS club names to abbreviated names
type Clubs map[string]string

var allClubs = Clubs{
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

func (c *Clubs) getKey(v string) (string, bool) {
	for key, value := range *c {
		if v == value {
			return key, true
		}
	}
	return "", false
}

// HasVal returns true if s is the full or abbreviated name of a club
func (c *Clubs) HasVal(s string) bool {
	if _, ok := (*c)[s]; ok {
		return true
	}
	_, ok := (*c).getKey(s)
	return ok
}

// Abv returns the abbreviated name of a club
func (c *Clubs) Abv(s string) string {
	if abv, ok := (*c)[s]; ok {
		return abv
	}
	if _, ok := (*c).getKey(s); ok {
		return s
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

var allDPs = `
	Fanendo Adi,
	Romain Alessandrini,
	Miguel Almiron,
	Jozy Altidore,
	Paul Arriola,
	Ezequiel Barco,
	Alejandro Bedoya,
	Sebastian Blanco,
	Michael Bradley,
	Josue Colman,
	Cristian Colman,
	Yohan Croizet,
	Claude Dielna,
	Borek Dockal,
	Giovani dos Santos,
	Jonathan dos Santos,
	Dom Dwyer,
	Alberth Elis,
	Shkelzen Gashi,
	Sebastian Giovinco,
	Carlos Gruezo,
	Felipe Gutierrez,
	Federico Higuain,
	Andre Horta,
	Tim Howard,
	Sacha Kljestan,
	Nicolas Lodeiro,
	Josef Martinez,
	Tomas Martinez,
	Jesus Medina,
	Lucas Melano,
	Maxi Moralez,
	Santiago Mosquera,
	Nemanja Nikolic,
	Ignacio Piatti,
	Valeri "Vako" Qazaishvili,
	Darwin Quintero,
	Angelo Rodriguez,
	Wayne Rooney,
	Alejandro 'Kaku' Romero,
	Diego Rossi,
	Raul Ruidiaz,
	Johnny Russell,
	Albert Rusnak,
	Pedro Santos,
	Jefferson Savarino,
	Bastian Schweinsteiger,
	Brek Shea,
	Saphir Taider,
	Milton Valenzuela,
	Diego Valeri,
	Carlos Vela,
	David Villa,
	Kendall Waston,
	Chris Wondolowski,
	Bradley Wright-Phillips
`

// Pos is the set of player positions
type Pos []string

var allPos = Pos{"F", "M-F", "F-M", "F/M", "GK", "D", "D-M", "M-D", "M", "M/F"}

// HasVal returns true if s is in p
func (p *Pos) HasVal(s string) bool {
	s = strings.ToUpper(s)
	for _, pos := range *p {
		if pos == s {
			return true
		}
	}
	return false
}

// Set sets the value of p from a comma separated list of positions
func (p *Pos) Set(s string) error {
	for _, pos := range strings.Split(s, ",") {
		pos = strings.ToUpper(strings.TrimSpace(pos))
		if !allPos.HasVal(pos) {
			return fmt.Errorf("valid values: %s", allPos.String())
		}
		*p = append(*p, pos)
	}
	return nil
}

func (p *Pos) String() string { return strings.Join(*p, ", ") }

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
		all        Players
		clubs      Clubs
		players    Players
		pos        Pos
		clubTotals = make(ClubTotals, len(allClubs))
	)
	log.SetFlags(0)
	flag.Var(&clubs, "clubs", "comma separated list of mls clubs")
	flag.Var(&players, "players", "comma separated list of mls players")
	flag.Var(&pos, "pos", "comma separated list of player positions")
	club := flag.Bool("sort", true, "sort by club")
	dp := flag.Bool("dp", false, "only show DP players")
	data := flag.String("data", "2018_09_15_data", "data file")
	flag.BoolVar(&Debug, "debug", false, "print data lines that don't match")
	flag.Parse()

	if _, err := os.Stat(*data); err != nil {
		if e, ok := err.(*os.PathError).Err.(syscall.Errno); ok && e == 2 {
			// no such file or dir
			if *data, ok = dataFromSource(*data); !ok {
				log.Fatal("unable to read data:", err)
			}
		} else {
			log.Fatal("unable to read data:", err)
		}
	}

	dps := Players{}
	_ = dps.Set(allDPs)

	f, err := os.Open(*data)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		sep := " "
		if strings.Contains(scanner.Text(), "\t") {
			sep = "\t"
		}
		tokens := strings.Split(scanner.Text(), sep)
		player := Player{}
		for _, token := range tokens {
			if token == "" {
				continue
			}
			switch {
			case allClubs.HasVal(token):
				player.Club = allClubs.Abv(token)

			case allPos.HasVal(token):
				player.Pos = token

			case token[0] == '$', token[0] >= '0' && token[0] <= '9':
				token = strings.TrimLeft(token, "$")
				if token == "" {
					continue
				}
				val, err := strconv.ParseFloat(strings.Replace(token, ",", "", -1), 32)
				if err != nil {
					continue
				}
				if player.BaseSalary == 0 {
					player.BaseSalary = val
				} else {
					player.Compensation = val
				}

			default:
				if player.Name == "" {
					player.Name = token
				} else {
					player.Name += " " + token
				}
			}
		}
		if player.Club == "" && player.Pos == "" && player.Compensation < 30000.00 {
			Dprintln("no match:", player)
			continue
		}
		if clubs != nil && !clubs.HasVal(player.Club) {
			continue
		}
		if pos != nil && !pos.HasVal(player.Pos) {
			continue
		}
		if *dp && !dps.HasVal(player.Name) {
			continue
		}
		if players != nil && !players.HasVal(player.Name) {
			continue
		}
		if player.Club == "" {
			Dprintln("no club", player)
		}
		if player.Pos == "" {
			Dprintln("no pos", player)
		}
		if player.Compensation < 30000.00 {
			Dprintln("no compensation", player)
		}

		all = append(all, player)
		clubTotals[player.Club] += player.Compensation
	}

	if len(all) == 0 {
		fmt.Println("No matches found")
		return
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Compensation > all[j].Compensation })
	if *club {
		sort.SliceStable(all, func(i, j int) bool { return all[i].Club < all[j].Club })
	}

	i := 1
	lastClub := all[0].Club
	for _, data := range all {
		if *club && data.Club != lastClub {
			i = 1
			lastClub = data.Club
			Printf("\n")
		}
		Printf("%-3d %-5s %-3s %-25s: %s\n", i, data.Club, data.Pos, data.Name, commaf(data.Compensation))
		i++
	}

	Printf("\n\n")
	for i, v := range clubTotals.Sort() {
		Printf("%-2d %-5s total: %s\n", i+1, v.Key, commaf(v.Value))
	}

	Dprintln()
	for _, n := range dps {
		if !all.HasVal(n.Name) {
			Dprintln("dp not found:", n.Name)
		}
	}
}

func dataFromSource(data string) (string, bool) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}
	path := filepath.Join(filepath.Dir(f), "data", data)
	fi, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	return path, fi.Mode().IsRegular()
}
