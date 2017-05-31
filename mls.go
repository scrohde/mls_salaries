package main

import (
	"bufio"
	"bytes"
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

var clubs = map[string]interface{}{
	"NE":    nil,
	"ORL":   nil,
	"SJ":    nil,
	"VAN":   nil,
	"CLB":   nil,
	"DC":    nil,
	"MNUFC": nil,
	"SEA":   nil,
	"CHI":   nil,
	"COL":   nil,
	"DAL":   nil,
	"KC":    nil,
	"LA":    nil,
	"LAFC":  nil,
	"MTL":   nil,
	"NYRB":  nil,
	"TOR":   nil,
	"ATL":   nil,
	"HOU":   nil,
	"NYCFC": nil,
	"PHI":   nil,
	"POR":   nil,
	"RSL":   nil,
}

func main() {
	var all []mlsData
	f, err := os.Open("2017_04_15_data")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		input := scanner.Text()
		tokens := strings.Split(input, " ")
		if _, ok := clubs[tokens[0]]; !ok {
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
			break
		}
	}

	sort.SliceStable(all, func(i, j int) bool { return all[i].Compensation > all[j].Compensation })
	sort.SliceStable(all, func(i, j int) bool { return all[i].Club < all[j].Club })
	for _, data := range all {
		fmt.Printf("%+v\n", data)
	}

	for k, _ := range clubs {
		total := 0.0
		for _, data := range all {
			if data.Club == k {
				total += data.Compensation
			}
		}
		//fmt.Printf("%5s:  %.2f\n", k, total)
		fmt.Printf("%5s:  %20s\n", k, Commaf(total))
	}
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
