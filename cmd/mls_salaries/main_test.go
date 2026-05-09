package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) string {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Env = append(os.Environ(), "GOCACHE=/tmp/go-build-cache")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run failed: %v\n%s", err, out)
	}
	return string(out)
}

func TestCLIParses2025SpaceDelimitedRows(t *testing.T) {
	out := runCLI(t, "-data", "2025_06_25_data", "-players", "Luis Abram", "-sort=false")

	if !strings.Contains(out, "ATL") {
		t.Fatalf("expected parsed Atlanta club abbreviation in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Center-back") {
		t.Fatalf("expected parsed position in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Luis Abram") {
		t.Fatalf("expected player name in output, got:\n%s", out)
	}
	if strings.Contains(out, "Luis Abram Atlanta United") {
		t.Fatalf("club name leaked into parsed player name:\n%s", out)
	}
}

func TestCLIPositionFilterMatchesAbbreviatedData(t *testing.T) {
	out := runCLI(t, "-data", "2013_09_15_data", "-pos", "F", "-sort=false")

	if strings.Contains(out, "No matches found") {
		t.Fatalf("expected abbreviated position filter to match old data, got:\n%s", out)
	}
	if !strings.Contains(out, "Agudelo Juan") {
		t.Fatalf("expected a forward from the 2013 data, got:\n%s", out)
	}
}
