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
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

//go:embed data/*
var dataFS embed.FS

// usage prints usage information and lists available data files.
func usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	files, err := fs.Glob(dataFS, "data/*_data")
	checkErr(err)
	if len(files) > 0 {
		// If there's an odd number of files, append a dummy value to maintain pairs.
		if len(files)%2 != 0 {
			files = append(files, "data/")
		}
		fmt.Printf("\nData files:\n")
		for i := 0; i < len(files); i += 2 {
			fmt.Printf("  %s, %s\n", files[i][len("data/"):], files[i+1][len("data/"):])
		}
	}
}

func main() {
	flag.Usage = usage
	var (
		// playersData holds the filtered player records.
		playersData Players

		// clubs, filterPlayers, and pos are command-line filters.
		clubs         Clubs
		filterPlayers Players
		pos           Pos

		sortByClub = flag.Bool("sort", true, "sort by club")
		dataFile   = flag.String("data", "2024_04_25_data", "data file")
		debug      = flag.Bool("debug", false, "print data lines that don't match")
		dps        = flag.Bool("dp", false, "players making above the maximum Targeted Allocation Money amount")
		clubTotals = make(ClubTotals, len(allClubs))
	)
	log.SetFlags(0)
	flag.Var(&clubs, "clubs", "comma separated list of MLS clubs")
	flag.Var(&filterPlayers, "players", "comma separated list of MLS players")
	flag.Var(&pos, "pos", "comma separated list of player positions")
	flag.Parse()

	// debugln prints debug output when the debug flag is set.
	debugln := func(a ...interface{}) {
		if *debug {
			fmt.Println(a...)
		}
	}

	// Open the data file, trying the local filesystem first, then the embedded FS.
	var r *bufio.Reader
	f, err := os.Open(*dataFile)
	if err != nil {
		fsFile, err := dataFS.Open("data/" + *dataFile)
		if err != nil {
			log.Fatal(err)
		}
		r = bufio.NewReader(fsFile)
	} else {
		r = bufio.NewReader(f)
	}

	// Determine the separator: if the first byte is a tab, use tab as the separator; otherwise, use a space.
	sep := " "
	b, err := r.ReadByte()
	checkErr(err)
	if b == '\t' {
		sep = "\t"
	} else {
		checkErr(r.UnreadByte())
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), sep)
		player := Player{}
		position := Position("")
		for _, token := range tokens {
			if token == "" {
				continue
			}
			switch {
			// Check if the token matches a club.
			case allClubs.HasVal(token):
				player.Club = allClubs.Abv(token)
			// Check if the token matches a position.
			case allPos.HasVal(token):
				player.Pos = token
			// If the token starts with '$' or a digit, treat it as a salary value.
			case token[0] == '$' || (token[0] >= '0' && token[0] <= '9'):
				token = strings.TrimLeft(token, "$")
				if token == "" {
					continue
				}
				// Parse salary as float64.
				val, err := strconv.ParseFloat(strings.Replace(token, ",", "", -1), 64)
				if err != nil {
					continue
				}

				if player.BaseSalary == 0 {
					player.BaseSalary = val
				} else {
					player.Compensation = val
				}
			// Otherwise, assume the token is part of the player's name.
			default:
				if player.Name == "" {
					player.Name = token
				} else {
					player.Name += " " + token
				}
			}
		}
		// Skip lines with insufficient data.
		if player.Club == "" && player.Pos == "" && player.Compensation < 30000.00 {
			debugln("no match:", player)
			continue
		}
		// Apply club, position, and player name filters if specified.
		if len(clubs) > 0 && !clubs.HasVal(player.Club) {
			continue
		}
		if len(pos) > 0 && !pos.HasVal(player.Pos) {
			continue
		}
		if len(filterPlayers) > 0 && !filterPlayers.HasVal(player.Name) {
			continue
		}
		// Filter for designated players if requested.
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

		playersData = append(playersData, player)
		clubTotals[player.Club] += player.Compensation
	}
	// Check for scanning errors.
	checkErr(scanner.Err())

	if len(playersData) == 0 {
		fmt.Println("No matches found")
		return
	}

	// Sort playersData by compensation in descending order.
	sort.Slice(playersData, func(i, j int) bool {
		return playersData[i].Compensation > playersData[j].Compensation
	})
	// Then group by club if requested.
	if *sortByClub {
		sort.SliceStable(playersData, func(i, j int) bool {
			return playersData[i].Club < playersData[j].Club
		})
	}

	var w io.Writer
	if !*debug {
		w = os.Stdout
	} else {
		w = io.Discard
	}
	t := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	i := 1
	lastClub := playersData[0].Club
	for _, data := range playersData {
		// Insert a blank line and reset the counter when the club changes.
		if *sortByClub && data.Club != lastClub {
			i = 1
			lastClub = data.Club
			_, err := fmt.Fprintln(t)
			checkErr(err)
		}
		_, err := fmt.Fprintf(t, "%d\t%s\t%s\t%s\t%s\n", i, data.Club, data.Pos, data.Name, commaf(data.Compensation))
		checkErr(err)
		i++
	}

	_, err = fmt.Fprintf(t, "\n\n")
	checkErr(err)

	// Assuming clubTotals.Sort returns a sorted slice of club totals with fields Key and Value.
	for i, v := range clubTotals.Sort() {
		_, err := fmt.Fprintf(t, "%d\t%s\ttotal: %s\n", i+1, v.Key, commaf(v.Value))
		checkErr(err)
	}
	checkErr(t.Flush())
	debugln()
}

// commaf returns a string representation of v with commas inserted for thousands.
// For example, 1234567.89 becomes "1,234,567.89".
func commaf(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.WriteByte('-')
		v = -v
	}

	// Format the float with two decimal places.
	s := strconv.FormatFloat(v, 'f', 2, 64)
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	fractionalPart := ""
	if len(parts) > 1 {
		fractionalPart = parts[1]
	}

	n := len(integerPart)
	remainder := n % 3
	if remainder > 0 {
		buf.WriteString(integerPart[:remainder])
		if n > remainder {
			buf.WriteByte(',')
		}
	}
	for i := remainder; i < n; i += 3 {
		buf.WriteString(integerPart[i : i+3])
		if i+3 < n {
			buf.WriteByte(',')
		}
	}
	if fractionalPart != "" {
		buf.WriteByte('.')
		buf.WriteString(fractionalPart)
	}
	return buf.String()
}

// checkErr is a helper function that logs a fatal error if err is non-nil.
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
