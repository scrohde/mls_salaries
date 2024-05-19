package main

import (
	"bufio"
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

//go:embed data/*
var dataFS embed.FS

func usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	files, err := fs.Glob(dataFS, "data/*_data")
	check(0, err)
	if len(files) > 0 {
		if len(files)%2 != 0 {
			files = append(files, "data/")
		}
		fmt.Printf("\ndata files: \n")
		for i := 0; i < len(files); i += 2 {
			fmt.Printf("  %s, %s\n", files[i][len("data/"):], files[i+1][len("data/"):])
		}
	}
}

func main() {
	flag.Usage = usage
	var (
		all        Players
		clubs      Clubs
		players    Players
		pos        Pos
		sortByClub = flag.Bool("sort", true, "sort by club")
		data       = flag.String("data", "2024_04_25_data", "data file")
		debug      = flag.Bool("debug", false, "print data lines that don't match")
		dps        = flag.Bool("dp", false, "players making above the maximum Targeted Allocation Money amount")
		clubTotals = make(ClubTotals, len(allClubs))
	)
	log.SetFlags(0)
	flag.Var(&clubs, "clubs", "comma separated list of mls clubs")
	flag.Var(&players, "players", "comma separated list of mls players")
	flag.Var(&pos, "pos", "comma separated list of player positions")
	flag.Parse()

	debugln := func(a ...interface{}) {
		if *debug {
			fmt.Println(a...)
		}
	}

	var r *bufio.Reader
	f, err := os.Open(*data)
	if err != nil {
		f, err := dataFS.Open("data/" + *data)
		if err != nil {
			log.Fatal(err)
		}
		r = bufio.NewReader(f)
	} else {
		r = bufio.NewReader(f)
	}

	var sep = " "
	if b, _ := r.ReadByte(); string(b) == "\t" {
		sep = "\t"
	} else {
		_ = r.UnreadByte()
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
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
			debugln("no match:", player)
			continue
		}
		if clubs != nil && !clubs.HasVal(player.Club) {
			continue
		}
		if pos != nil && !pos.HasVal(player.Pos) {
			continue
		}
		if players != nil && !players.HasVal(player.Name) {
			continue
		}
		if *dps && player.Compensation < 1_612_500 {
			continue
		}
		if player.Club == "" {
			debugln("no club", player)
		}
		if player.Pos == "" {
			debugln("no pos", player)
		}
		if player.Compensation < 30000.00 {
			debugln("no compensation", player)
		}

		all = append(all, player)
		clubTotals[player.Club] += player.Compensation
	}

	if len(all) == 0 {
		fmt.Println("No matches found")
		return
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Compensation > all[j].Compensation })
	if *sortByClub {
		sort.SliceStable(all, func(i, j int) bool { return all[i].Club < all[j].Club })
	}
	var w io.Writer
	if !*debug {
		w = os.Stdout
	} else {
		w = io.Discard
	}
	t := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i := 1
	lastClub := all[0].Club
	for _, data := range all {
		if *sortByClub && data.Club != lastClub {
			i = 1
			lastClub = data.Club
			check(fmt.Fprintln(t))
		}
		check(fmt.Fprintf(t, "%d\t%s\t%s\t%s\t%s\n", i, data.Club, data.Pos, data.Name, commaf(data.Compensation)))
		i++
	}

	check(fmt.Fprintf(t, "\n\n"))
	for i, v := range clubTotals.Sort() {
		check(fmt.Fprintf(t, "%d\t%s\ttotal: %s\n", i+1, v.Key, commaf(v.Value)))
	}
	err = t.Flush()
	if err != nil {
		log.Fatal(err)
	}
	debugln()
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

func check(_ interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
}
