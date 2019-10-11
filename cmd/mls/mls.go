package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"golang.org/x/xerrors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Player is an MLS player
type Player struct {
	Club         string
	Name         string
	Pos          string
	Goals        int
	Assists      int
	Compensation float64
}

func main() {
	var r *csv.Reader
	var players []Player

	filename := "ASAshootertable.csv"
	if path, ok := dataFromSource(filename); !ok {
		fmt.Printf("%+v", xerrors.Errorf("unable ot find data file: %s: %w", filename))
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
		13: xPlace 14: G-xG 15: KeyP 16: Dist.key 17: A 18: xA 19: A-xA 20: xG+xA 21: PA 22: xPA 23: xG/shot 24: xA/pass
		25: G-xG/shot 26: A-xA/pass 27: Comp ($K) 28: Team/96 29: Min/96 30: Pos/96 31: Shots/96 32: SoT/96 33: G/96
		34: xG/96 35: xPlace/96 36: G-xG/96 37: KeyP/96 38: A/96 39: xA/96 40: A-xA/96 41: xG+xA/96 42: PA/96 43: xPA/96
		44: Comp ($K)/96 45: extreme1 46: extreme2 47: plotnames
		*/
		p := Player{
			Club:         record[3],
			Name:         record[2],
			Pos:          record[6],
			Goals:        goals,
			Assists:      assists,
			Compensation: comp,
		}
		players = append(players, p)
	}

	sort.Slice(players, func(i, j int) bool { return players[i].Compensation > players[j].Compensation })
	sort.SliceStable(players, func(i, j int) bool { return players[i].Goals+players[i].Assists > players[j].Goals+players[j].Assists })

	w := os.Stdout
	t := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for i, data := range players {
		combined := data.Goals + data.Assists
		per := data.Compensation/float64(combined)
		_, err := fmt.Fprintf(t, "%d\t%s\t%s\t%d/%d\t%s\t%s\t(%s)\n", i, data.Club, data.Pos, data.Goals, data.Assists, data.Name, commaf(data.Compensation), commaf(per))
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
	path := filepath.Join(filepath.Dir(f)+"../../..", "data", data)
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
