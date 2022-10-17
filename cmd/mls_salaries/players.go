package main

import (
	"fmt"
	"strings"
)

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

// HasVal returns true if any players name contains s
func (p *Players) HasVal(val string) bool {
	for _, player := range *p {
		if strings.Contains(strings.ToLower(val), strings.ToLower(player.Name)) {
			return true
		}
	}
	return false
}

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
