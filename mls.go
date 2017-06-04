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
	Club         string
	Name         string
	Pos          string
	BaseSalary   float64
	Compensation float64
}

type Clubs []string

func (c *Clubs) Set(s string) error {
	*c = strings.Split(s, ",")
	for i, club := range *c {
		(*c)[i] = strings.TrimSpace(strings.ToUpper(club))
		if !allClubs.hasVal((*c)[i]) {
			return fmt.Errorf("invalid club: %s\nvalid clubs: %s", club, allClubs.String())
		}
	}
	return nil
}

func (c *Clubs) String() string {
	return strings.Join(*c, ", ")
}

func (c *Clubs) hasVal(s string) bool {
	for _, val := range *c {
		if val == s {
			return true
		}
	}
	return false
}

var allClubs = Clubs{
	"NE",
	"ORL",
	"SJ",
	"VAN",
	"CLB",
	"DC",
	"MNUFC",
	"SEA",
	"CHI",
	"COL",
	"DAL",
	"KC",
	"LA",
	"LAFC",
	"MTL",
	"NYRB",
	"TOR",
	"ATL",
	"HOU",
	"NYCFC",
	"PHI",
	"POR",
	"RSL",
}

// A data structure to hold a key/value pairs.
type KeyValue struct {
	Key   string
	Value float64
}

// sortMapByValue turns a map into a PairList, then sorts and returns it.
func sortMapByValue(m map[string]float64) []KeyValue {
	p := make([]KeyValue, len(m))
	i := 0
	for k, v := range m {
		p[i] = KeyValue{k, v}
		i++
	}
	sort.Slice(p, func(i, j int) bool { return p[i].Value > p[j].Value })
	return p
}

func Commaf(v float64) string {
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
	var club = flag.Bool("club", true, "sort by club")
	flag.Parse()

	if len(clubs) == 0 {
		clubs = allClubs
	}

	f, err := os.Open("2017_04_15_data")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)

	clubTotals := make(map[string]float64, 30)
	for scanner.Scan() {
		input := scanner.Text()
		tokens := strings.Split(input, " ")
		if !clubs.hasVal(tokens[0]) {
			continue
		}

		for i, val := range tokens[1:] {
			if val != "$" {
				continue
			}
			if tokens[i+3] != "$" {
				break
			}
			data := mlsData{}
			data.Club = tokens[0]
			data.Pos = tokens[i]
			data.Name = strings.Join(tokens[1:i], " ")
			data.BaseSalary, err = strconv.ParseFloat(strings.Replace(tokens[i+2], ",", "", -1), 32)
			data.Compensation, err = strconv.ParseFloat(strings.Replace(tokens[i+4], ",", "", -1), 32)
			all = append(all, data)

			clubTotals[data.Club] += data.Compensation
			break
		}
	}

	sort.Slice(all, func(i, j int) bool { return all[i].Compensation > all[j].Compensation })
	if *club {
		sort.SliceStable(all, func(i, j int) bool { return all[i].Club < all[j].Club })
	}

	for _, data := range all {
		fmt.Printf("%-5s %-25s: %s\n", data.Club, data.Name, Commaf(data.Compensation))
	}

	fmt.Print("\n\n")
	p := sortMapByValue(clubTotals)
	for _, v := range p {
		fmt.Printf("%-5s total: %s\n", v.Key, Commaf(v.Value))
	}
}
