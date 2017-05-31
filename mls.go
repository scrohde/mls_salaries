package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type mlsData struct {
	Club         string
	LastName     string
	FirstName    string
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

func mlsSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	adv := 0
	for i := 0; i < 10; i++ {
		advance, token, err = bufio.ScanLines(data[adv:], atEOF)
		if err != nil {
			return
		}
		if atEOF && len(data[adv:]) == 0 {
			return
		}

		if bytes.Index(token, []byte("$")) == -1 {
			adv += advance
			continue
		}

		datum := bytes.Split(token, []byte(" "))
		if _, ok := clubs[string(datum[0])]; !ok {
			adv += advance
			continue
		} else {
			adv += advance
			break
		}
	}
	return adv, token, nil
}

func main() {

	var all []mlsData
	f, err := os.Open("2017_04_15_data")
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(mlsSplit)

	for scanner.Scan() {
		data := mlsData{}
		input := scanner.Text()
		fmt.Printf("input = %+v\n", input)

		tokens := strings.Split(input, " ")
		data.Club = tokens[0]
		data.LastName = tokens[1]
		// we will have a variable # of names
		for i := range tokens[2:] {
			if tokens[i] == "$" {
				data.Pos = tokens[i-1]
				if i > 3 {
					data.FirstName = strings.Join(tokens[2:i], " ")
				}
				data.BaseSalary, err = strconv.ParseFloat(strings.Replace(tokens[i+1], ",", "", -1), 32)
				data.Compensation, err = strconv.ParseFloat(strings.Replace(tokens[i+3], ",", "", -1), 32)
			}
		}

		all = append(all, data)
	}
	for _, a := range all {
		fmt.Printf("a = %+v\n", a)
	}
}
