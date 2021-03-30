package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Set struct{}

func (s Set) Contains(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

func transformDPs(s string) []string {
	var result []string
	var set Set

	t := transform.Chain(norm.NFD, runes.Remove(set), norm.NFC)
	names := strings.Split(s, "\n")
	for _, n := range names {
		name := strings.Split(n, ",")
		if len(name) != 2 {
			continue
		}
		first, _, _ := transform.String(t, name[1])
		first = strings.TrimSpace(strings.TrimRight(first, "*"))
		last, _, _ := transform.String(t, name[0])
		last = strings.TrimSpace(strings.TrimRight(last, "*"))
		result = append(result, fmt.Sprintf("%s %s", first, last))
	}
	return result
}

func main() {
	b, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for _, name := range transformDPs(string(b)) {
		fmt.Printf("\t%s,\n", name)
	}
}
