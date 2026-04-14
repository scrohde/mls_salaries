package main

import (
	"errors"
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
	names := make([]string, 0, len(*p))
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

type Position string

func (p *Position) String() string {
	return string(*p)
}

func (p *Position) Set(posStr string) error {
	switch strings.ToUpper(strings.TrimSpace(posStr)) {
	case "CENTER FORWARD", "F", "FORWARD":
		*p = "Center Forward"
	case "LEFT WING", "LW":
		*p = "Left Wing"
	case "RIGHT WING", "RW":
		*p = "Right Wing"
	case "ATTACKING MIDFIELD", "F-M", "F/M", "M-F", "M/F", "AM":
		*p = "Attacking Midfield"
	case "CENTRAL MIDFIELD", "M", "MF", "MIDFIELDER", "CM":
		*p = "Central Midfield"
	case "DEFENSIVE MIDFIELD", "D-M", "M-D", "D/M", "M/D", "DM":
		*p = "Defensive Midfield"
	case "LEFT MIDFIELD", "LM":
		*p = "Left Midfield"
	case "RIGHT MIDFIELD", "RM":
		*p = "Right Midfield"
	case "LEFT-BACK", "LB":
		*p = "Left-back"
	case "CENTER-BACK", "D", "DEFENDER", "CB":
		*p = "Center-back"
	case "RIGHT-BACK", "RB":
		*p = "Right-back"
	case "GOALKEEPER", "GK":
		*p = "Goalkeeper"
	case "SUBSTITUTE":
		*p = "Substitute"
	default:
		return errors.New("invalid position")
	}
	return nil
}

// Pos is the set of player positions used for filtering and validation
type Pos []string

// Set sets the value of p from a comma separated list of positions
func (p *Pos) Set(s string) error {
	v := new(Position)
	for _, pos := range strings.Split(s, ",") {
		if err := v.Set(pos); err != nil {
			return err
		}
		*p = append(*p, strings.ToUpper(v.String()))
	}
	return nil
}

// HasVal returns true if val matches any position in p (case-insensitive)
func (p Pos) HasVal(val string) bool {
	val = strings.ToUpper(strings.TrimSpace(val))
	for _, pos := range p {
		if strings.ToUpper(pos) == val {
			return true
		}
	}
	return false
}

func (p Pos) String() string { return strings.Join(p, ", ") }

// allPos enumerates valid position tokens
var allPos = Pos{"F", "M-F", "F-M", "F/M", "GK", "D", "D-M", "M-D", "M", "M/F",
	"RIGHT WING", "CENTER-BACK", "DEFENSIVE MIDFIELD", "RIGHT WING", "CENTRAL MIDFIELD", "CENTER FORWARD", "RIGHT-BACK",
	"ATTACKING MIDFIELD", "GOALKEEPER", "LEFT-BACK", "LEFT WING", "RIGHT MIDFIELD", "LEFT MIDFIELD", "MIDFIELDER", "FORWARD", "DEFENDER"}
