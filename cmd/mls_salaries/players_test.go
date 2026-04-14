package main

import "testing"

func TestPlayersString(t *testing.T) {
	var p Players
	if err := p.Set("Lionel Messi,Neymar"); err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	got := p.String()
	const want = "Lionel Messi, Neymar"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
