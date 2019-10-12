package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/xerrors"
)

// Player is an MLS player
type Player struct {
	Club         string
	Name         string
	Pos          string
	Goals        int
	Assists      int
	Compensation float64
	GAPerDollar  float64
}

type Clubs []string

func (c *Clubs) Set(v string) error {
	clubs := strings.Split(v, ",")
	for _, club := range clubs {
		club = strings.ToUpper(strings.TrimSpace(club))
		if allClubs.Has(club) {
			*c = append(*c, club)
		} else {
			return fmt.Errorf("valid clubs: %s", allClubs)
		}
	}
	return nil
}
func (c *Clubs) Has(v string) bool {
	for _, club := range *c {
		if v == club {
			return true
		}
	}
	return false
}
func (c *Clubs) String() string {
	if c == nil {
		return ""
	}
	return strings.Join(*c, ", ")
}

var allClubs = Clubs{
	"COL",
	"LAG",
	"MIN",
	"ORL",
	"CLB",
	"PHI",
	"CHI",
	"HOU",
	"FCD",
	"POR",
	"RSL",
	"CIN",
	"SJE",
	"SKC",
	"VAN",
	"NER",
	"DCU",
	"ATL",
	"LAFC",
	"TOR",
	"NYC",
	"NYRB",
	"SEA",
	"MTL",
}

func main() {
	var (
		r       *csv.Reader
		players []Player
		clubs   = &Clubs{}
	)

	flag.Var(clubs, "clubs", "comma separated list of clubs")
	flag.Parse()

	filename := "ASAshootertable.csv"
	if path, ok := dataFromSource(filename); !ok {
		//fmt.Printf("%+v", xerrors.Errorf("unable ot find data file: %s", filename))
		fmt.Printf("%+v", xerrors.Errorf("unable ot find data file: %s", filename))
		os.Exit(1)
	} else {
		f, err := os.Open(path)
		check(err)
		r = csv.NewReader(f)
	}
	_, err := r.Read()
	//for i, title := range titles {
	//	fmt.Printf("%d: %s\n", i, title)
	//}
	check(err)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)

		if len(*clubs) != 0 {
			if !clubs.Has(record[3]) {
				continue
			}
		}
		comp, err := strconv.ParseFloat(record[27], 32)
		if err != nil {
			comp = 0
		}
		comp = comp * 1000
		goals, err := strconv.Atoi(record[11])
		if err != nil {
			goals = 0
		}
		assists, err := strconv.Atoi(record[17])
		if err != nil {
			assists = 0
		}
		/*
			0: First 1: Last 2: Player 3: Team 4: Season 5: Min 6: Pos 7: Shots 8: SoT 9: Dist 10: Solo 11: G 12: xG
			13: xPlace 14: G-xG 15: KeyP 16: Dist.key 17: A 18: xA 19: A-xA 20: xG+xA 21: PA 22: xPA 23: xG/shot
			24: xA/pass 25: G-xG/shot 26: A-xA/pass 27: Comp ($K) 28: Team/96 29: Min/96 30: Pos/96 31: Shots/96
			32: SoT/96 33: G/96 34: xG/96 35: xPlace/96 36: G-xG/96 37: KeyP/96 38: A/96 39: xA/96 40: A-xA/96
			41: xG+xA/96 42: PA/96 43: xPA/96 44: Comp ($K)/96 45: extreme1 46: extreme2 47: plotnames
		*/
		p := Player{
			Club:         record[3],
			Name:         record[2],
			Pos:          record[6],
			Goals:        goals,
			Assists:      assists,
			Compensation: comp,
			GAPerDollar:  comp / float64(goals+assists),
		}
		players = append(players, p)
	}

	dollars := []float64{}
	var median float64
	for _, p := range players {
		if p.GAPerDollar > 0 && p.Pos != "CDM" && p.Pos != "CB" && p.Pos != "GK" {
			dollars = append(dollars, p.GAPerDollar)
		}
	}
	sort.Float64s(dollars)
	half := len(dollars) / 2
	if len(dollars)%2 != 0 {
		// odd
		median = (dollars[half-1] + dollars[half]) / 2
	} else if half != 0 {
		median = dollars[half]
	}
	fmt.Println("median dollars per goals+assists:", commaf(median))
	sort.Slice(players, func(i, j int) bool { return players[i].Compensation > players[j].Compensation })
	sort.SliceStable(players, func(i, j int) bool { return players[i].Goals+players[i].Assists > players[j].Goals+players[j].Assists })
	sort.SliceStable(players, func(i, j int) bool {
		return players[i].GAPerDollar < players[j].GAPerDollar
	})

	w := os.Stdout
	t := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, data := range players {
		_, err := fmt.Fprintf(t, "%d\t%s\t%s\t%d/%d\t%s\t%s\t(%s)\n", i, data.Club, data.Pos, data.Goals, data.Assists, data.Name, commaf(data.Compensation), commaf(data.GAPerDollar))
		check(err)
	}
	check(t.Flush())
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

func dataFromSource(data string) (string, bool) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}
	path := filepath.Join(filepath.Dir(f), "../..", "data", data)
	fi, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	return path, fi.Mode().IsRegular()
}

func check(err error) {
	if err != nil {
		log.Printf("%+v", err)
		os.Exit(1)
	}
}
