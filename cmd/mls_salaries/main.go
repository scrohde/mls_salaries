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

func tokenizeLine(line, sep string) []string {
	if sep == "\t" {
		raw := strings.Split(line, sep)
		tokens := make([]string, 0, len(raw))
		for _, token := range raw {
			token = strings.TrimSpace(token)
			if token != "" {
				tokens = append(tokens, token)
			}
		}
		return tokens
	}
	return strings.Fields(line)
}

func matchWords(tokens []string, start int, words []string) bool {
	if start+len(words) > len(tokens) {
		return false
	}
	for i, word := range words {
		if !strings.EqualFold(tokens[start+i], word) {
			return false
		}
	}
	return true
}

func matchClubAt(tokens []string, start int) (string, int) {
	bestAbv := ""
	bestLen := 0
	for full, abv := range allClubs {
		switch {
		case strings.EqualFold(tokens[start], full), strings.EqualFold(tokens[start], abv):
			if bestLen < 1 {
				bestAbv = abv
				bestLen = 1
			}
		default:
			words := strings.Fields(full)
			if len(words) > bestLen && matchWords(tokens, start, words) {
				bestAbv = abv
				bestLen = len(words)
			}
		}
	}
	return bestAbv, bestLen
}

func matchPositionAt(tokens []string, start int) (string, int) {
	maxWords := 3
	if remaining := len(tokens) - start; remaining < maxWords {
		maxWords = remaining
	}
	for n := maxWords; n > 0; n-- {
		position := Position("")
		if position.Set(strings.Join(tokens[start:start+n], " ")) == nil {
			return position.String(), n
		}
	}
	return "", 0
}

func parseSalary(token string) (float64, bool) {
	if token == "" || (token[0] != '$' && (token[0] < '0' || token[0] > '9')) {
		return 0, false
	}
	token = strings.TrimLeft(token, "$")
	if token == "" {
		return 0, false
	}
	val, err := strconv.ParseFloat(strings.Replace(token, ",", "", -1), 64)
	return val, err == nil
}

func parsePlayerLine(line, sep string) Player {
	tokens := tokenizeLine(line, sep)
	player := Player{}
	for i := 0; i < len(tokens); {
		token := tokens[i]
		if token == "$" {
			i++
			continue
		}
		if club, n := matchClubAt(tokens, i); n > 0 {
			player.Club = club
			i += n
			continue
		}
		if pos, n := matchPositionAt(tokens, i); n > 0 {
			player.Pos = pos
			i += n
			continue
		}
		if val, ok := parseSalary(token); ok {
			if player.BaseSalary == 0 {
				player.BaseSalary = val
			} else {
				player.Compensation = val
			}
			i++
			continue
		}
		if player.Name == "" {
			player.Name = token
		} else {
			player.Name += " " + token
		}
		i++
	}
	return player
}

func hasPosition(positions Positions, pos string) bool {
	for _, position := range positions {
		if strings.EqualFold(position, pos) {
			return true
		}
	}
	return false
}

func main() {
	flag.Usage = usage
	var (
		// playersData holds the filtered player records.
		playersData Players

		// clubs, filterPlayers, and pos are command-line filters.
		clubs         Clubs
		filterPlayers Players
		pos           Positions

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
		player := parsePlayerLine(scanner.Text(), sep)
		// Skip lines with insufficient data.
		if player.Club == "" && player.Pos == "" && player.Compensation < 30000.00 {
			debugln("no match:", player)
			continue
		}
		// Apply club, position, and player name filters if specified.
		if len(clubs) > 0 && !clubs.HasVal(player.Club) {
			continue
		}
		if len(pos) > 0 && !hasPosition(pos, player.Pos) {
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
