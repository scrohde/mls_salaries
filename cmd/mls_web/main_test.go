package main

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProcessDataParses2025SpaceDelimitedRows(t *testing.T) {
	players, totals, err := processData("2025_06_25_data", "Atlanta United (ATL)", "Luis Abram", "Center-back", false, false)
	if err != nil {
		t.Fatalf("processData returned error: %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("expected one player, got %d: %#v", len(players), players)
	}

	player := players[0]
	if player.Club != "ATL" {
		t.Fatalf("expected club ATL, got %q", player.Club)
	}
	if player.Pos != "Center-back" {
		t.Fatalf("expected position Center-back, got %q", player.Pos)
	}
	if player.Name != "Luis Abram" {
		t.Fatalf("expected clean player name, got %q", player.Name)
	}
	if player.Compensation != 871888 {
		t.Fatalf("expected compensation 871888, got %f", player.Compensation)
	}
	if totals["ATL"] != 871888 {
		t.Fatalf("expected ATL total 871888, got %f", totals["ATL"])
	}
}

func TestUpdateFiltersOOBTargetsExistingDatalists(t *testing.T) {
	req := httptest.NewRequest("GET", "/updateFilters?data=2025_06_25_data", nil)
	rec := httptest.NewRecorder()

	updateFiltersHandler(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<datalist id="clubs-list" hx-swap-oob="true">`,
		`<datalist id="positions-list" hx-swap-oob="true">`,
		`<datalist id="players-list" hx-swap-oob="true">`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected response to contain %q, got:\n%s", want, body)
		}
	}
	if strings.Contains(body, `id="clubs-datalist"`) {
		t.Fatalf("response targets a missing wrapper instead of the existing datalist:\n%s", body)
	}
}
