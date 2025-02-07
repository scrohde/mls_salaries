package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

// =============================================================================
// Embedded Assets
// =============================================================================

//go:embed data/*
var dataFS embed.FS

// =============================================================================
// Helper Types and Functions
// =============================================================================

// DataFileEntry holds a file's actual value and a nicely formatted display string.
type DataFileEntry struct {
	Value   string
	Display string
}

// formatDataFileName removes a trailing "_data" and replaces underscores with spaces.
func formatDataFileName(file string) string {
	if strings.HasSuffix(file, "_data") {
		file = file[:len(file)-len("_data")]
	}
	return strings.ReplaceAll(file, "_", " ")
}

// Clubs is a map of full club names to abbreviated names.
type Clubs map[string]string

var allClubs = Clubs{
	"MLS Pool":               "MLS",
	"New England Revolution": "NE",
	"Orlando City SC":        "ORL",
	"San Jose Earthquakes":   "SJ",
	"Vancouver Whitecaps":    "VAN",
	"Columbus Crew":          "CLB",
	"DC United":              "DC",
	"Minnesota United":       "MNUFC",
	"Seattle Sounders FC":    "SEA",
	"Chicago Fire":           "CHI",
	"Colorado Rapids":        "COL",
	"FC Dallas":              "DAL",
	"Sporting Kansas City":   "KC",
	"LA Galaxy":              "LA",
	"LAFC":                   "LAFC",
	"CF Montreal":            "MTL",
	"Montreal":               "MTL",
	"New York Red Bulls":     "NYRB",
	"Toronto FC":             "TOR",
	"Atlanta United":         "ATL",
	"Houston Dynamo":         "HOU",
	"New York City FC":       "NYCFC",
	"Philadelphia Union":     "PHI",
	"Portland Timbers":       "POR",
	"Real Salt Lake":         "RSL",
	"FC Cincinnati":          "CIN",
	"NY":                     "NYRB",
	"Chivas USA":             "CHV",
	"Nashville SC":           "NSC",
	"Inter Miami":            "MIA",
	"Austin FC":              "AFC",
	"Charlotte FC":           "CLT",
	"Major League Soccer":    "MLS",
	"St. Louis SC":           "STL",
	"St. Louis City SC":      "STL",
	"San Diego FC":           "SDFC",
}

func (c Clubs) getKey(val string) (string, bool) {
	for key, value := range c {
		if strings.EqualFold(val, value) {
			return key, true
		}
	}
	return "", false
}

func (c Clubs) HasVal(val string) bool {
	_, ok := c[val]
	if ok {
		return true
	}
	_, ok = c.getKey(val)
	return ok
}

func (c Clubs) Abv(fullName string) string {
	if abv, ok := c[fullName]; ok {
		return abv
	}
	if _, ok := c.getKey(fullName); ok {
		return fullName
	}
	return ""
}

// Player holds an MLS player's details.
type Player struct {
	Club          string
	Name          string
	Pos           string
	BaseSalary    float64
	Compensation  float64
	FormattedComp string
}

// Players is a list of Player.
type Players []Player

// Pos holds a list of valid positions.
type Pos []string

var allPos = Pos{"F", "M-F", "F-M", "F/M", "GK", "D", "D-M", "M-D", "M", "M/F",
	"Right Wing", "CENTER-BACK", "DEFENSIVE MIDFIELD", "RIGHT WING", "CENTRAL MIDFIELD", "CENTER FORWARD", "RIGHT-BACK",
	"ATTACKING MIDFIELD", "GOALKEEPER", "LEFT-BACK", "LEFT WING", "RIGHT MIDFIELD", "RIGHT WING", "LEFT MIDFIELD",
	"MIDFIELDER", "FORWARD", "DEFENDER"}

func (p Pos) HasVal(s string) bool {
	s = strings.ToUpper(s)
	for _, pos := range p {
		if pos == s {
			return true
		}
	}
	return false
}

// ClubTotals maps club names to total compensation.
type ClubTotals map[string]float64

// KeyValue holds a club and its total.
type KeyValue struct {
	Key   string
	Value float64
}

func (ct ClubTotals) Sort() []KeyValue {
	p := make([]KeyValue, 0, len(ct))
	for k, v := range ct {
		p = append(p, KeyValue{k, v})
	}
	sort.Slice(p, func(i, j int) bool { return p[i].Value > p[j].Value })
	return p
}

func commaf(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.WriteByte('-')
		v = -v
	}
	s := strconv.FormatFloat(v, 'f', 2, 64)
	parts := strings.Split(s, ".")
	integerPart := parts[0]
	fractionalPart := ""
	if len(parts) > 1 {
		fractionalPart = parts[1]
	}
	n := len(integerPart)
	remainder := n % 3
	if remainder > 0 {
		buf.WriteString(integerPart[:remainder])
		if n > remainder {
			buf.WriteByte(',')
		}
	}
	for i := remainder; i < n; i += 3 {
		buf.WriteString(integerPart[i : i+3])
		if i+3 < n {
			buf.WriteByte(',')
		}
	}
	if fractionalPart != "" {
		buf.WriteByte('.')
		buf.WriteString(fractionalPart)
	}
	return buf.String()
}

// =============================================================================
// Data Processing
// =============================================================================

func processData(dataFile, clubsStr, playersStr, posStr string, sortByClub, dp bool) (Players, ClubTotals, error) {
	var playersData Players
	clubTotals := make(ClubTotals)

	// Parse club filter from comma-separated string (if any)
	clubsFilter := make(Clubs)
	if clubsStr != "" {
		for _, name := range strings.Split(clubsStr, ",") {
			name = strings.TrimSpace(name)
			// If the club string contains a parenthetical abbreviation, strip it off.
			if idx := strings.Index(name, "("); idx != -1 {
				name = strings.TrimSpace(name[:idx])
			}
			// Allow matching if the input is contained in the full name or abbreviation.
			for full, abv := range allClubs {
				lowerName := strings.ToLower(name)
				if strings.Contains(strings.ToLower(full), lowerName) || strings.Contains(strings.ToLower(abv), lowerName) {
					clubsFilter[full] = abv
				}
			}
		}
	}

	// For players, the hidden input supplies a comma-separated list.
	var playersFilter []string
	if playersStr != "" {
		for _, name := range strings.Split(playersStr, ",") {
			playersFilter = append(playersFilter, strings.TrimSpace(name))
		}
	}

	var posFilter Pos
	if posStr != "" {
		for _, pos := range strings.Split(posStr, ",") {
			p := strings.ToUpper(strings.TrimSpace(pos))
			if allPos.HasVal(p) {
				posFilter = append(posFilter, p)
			}
		}
	}

	// Open the data file (try local first, then embedded)
	var r *bufio.Reader
	f, err := os.Open(dataFile)
	if err != nil {
		var fsFile fs.File
		fsFile, err = dataFS.Open("data/" + dataFile)
		if err != nil {
			return nil, nil, err
		}
		r = bufio.NewReader(fsFile)
	} else {
		r = bufio.NewReader(f)
	}

	// Determine separator: use tab if the first byte is '\t'; otherwise, use space.
	sep := " "
	b, err := r.ReadByte()
	if err != nil {
		return nil, nil, err
	}
	if b == '\t' {
		sep = "\t"
	} else {
		if err := r.UnreadByte(); err != nil {
			return nil, nil, err
		}
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), sep)
		player := Player{}
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}
			switch {
			case allClubs.HasVal(token):
				for full, abv := range allClubs {
					if strings.EqualFold(token, full) || strings.EqualFold(token, abv) {
						player.Club = abv
						break
					}
				}
			case allPos.HasVal(token):
				player.Pos = strings.ToUpper(token)
			case token[0] == '$' || (token[0] >= '0' && token[0] <= '9'):
				token = strings.TrimLeft(token, "$")
				if token == "" {
					continue
				}
				val, err := strconv.ParseFloat(strings.ReplaceAll(token, ",", ""), 64)
				if err != nil {
					continue
				}
				if player.BaseSalary == 0 {
					player.BaseSalary = val
				} else {
					player.Compensation = val
				}
			default:
				if player.Name == "" {
					player.Name = token
				} else {
					player.Name += " " + token
				}
			}
		}
		if player.Club == "" && player.Pos == "" && player.Compensation < 30000.00 {
			continue
		}
		if len(clubsFilter) > 0 && !clubsFilter.HasVal(player.Club) {
			continue
		}
		if len(posFilter) > 0 && !posFilter.HasVal(player.Pos) {
			continue
		}
		if len(playersFilter) > 0 {
			matched := false
			for _, name := range playersFilter {
				if strings.Contains(strings.ToLower(player.Name), strings.ToLower(name)) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if dp && player.Compensation < 1612500 {
			continue
		}
		playersData = append(playersData, player)
		clubTotals[player.Club] += player.Compensation
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	sort.Slice(playersData, func(i, j int) bool {
		return playersData[i].Compensation > playersData[j].Compensation
	})
	if sortByClub {
		sort.SliceStable(playersData, func(i, j int) bool {
			return playersData[i].Club < playersData[j].Club
		})
	}
	for i := range playersData {
		playersData[i].FormattedComp = commaf(playersData[i].Compensation)
	}
	return playersData, clubTotals, nil
}

// =============================================================================
// Template Helpers
// =============================================================================

func add(a, b int) int {
	return a + b
}

// =============================================================================
// Templates
// =============================================================================

var indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>MLS Compensation Filter</title>
	<script src="https://unpkg.com/htmx.org@1.9.2"></script>
	<style>
	  body { font-family: sans-serif; margin: 2rem; }
	  .container { display: flex; justify-content: space-between; }
	  .filters { width: 45%; }
	  .results { width: 45%; }
	  label { display: block; margin-top: 1rem; }
	  input, select { padding: 0.5rem; font-size: 1rem; width: 100%; }
	  table { border-collapse: collapse; margin-top: 1rem; width: 100%; }
	  th, td { padding: 0.5rem; border: 1px solid #ccc; }
	  th { background-color: #f4f4f4; }
	  /* Styles for auto-complete tag containers */
	  #selected-players, #selected-clubs, #selected-pos {
	    margin-top: 5px; 
	    border: 1px solid #ccc; 
	    padding: 5px; 
	    display: flex; 
	    flex-wrap: wrap;
	  }
	  .tag {
	    margin: 2px;
	    padding: 5px;
	    border: 1px solid #ccc;
	    border-radius: 4px;
	    background: #eee;
	    display: flex;
	    align-items: center;
	  }
	  .tag button {
	    margin-left: 5px;
	    border: none;
	    background: transparent;
	    cursor: pointer;
	  }
	</style>
</head>
<body>
	<h1>MLS Compensation Filter</h1>
	<div class="container">
	  <div class="filters">
		<form hx-post="/filter" hx-target="#results" hx-swap="innerHTML" hx-trigger="change delay:500ms">
			<!-- Data File Selection: Display formatted names -->
			<label>Data File:
				<select name="data" id="data-select" hx-get="/players" hx-target="#players-list" hx-trigger="change">
					{{range $i, $f := .DataFiles}}
						<option value="{{$f.Value}}" {{if eq $i 0}}selected{{end}}>{{$f.Display}}</option>
					{{end}}
				</select>
			</label>
			
			<!-- Clubs Auto-Complete (now showing "Full Name (Abv)") -->
			<label>Clubs:</label>
			<div id="club-selector">
				<input type="text" id="club-input" list="clubs-list" placeholder="Type club name and select" />
				<datalist id="clubs-list">
					{{range .ClubsList}}
						<option value="{{.}}">
					{{end}}
				</datalist>
				<div id="selected-clubs"></div>
				<input type="hidden" name="clubs" id="clubs-hidden" value="">
			</div>
			
			<!-- Players Auto-Complete -->
			<label>Players:</label>
			<div id="player-selector">
				<input type="text" id="player-input" list="players-list" placeholder="Type player name and select" />
				<datalist id="players-list">
					{{range .PlayersList}}
						<option value="{{.}}">
					{{end}}
				</datalist>
				<div id="selected-players"></div>
				<input type="hidden" name="players" id="players-hidden" value="">
			</div>
			
			<!-- Positions Auto-Complete -->
			<label>Positions:</label>
			<div id="pos-selector">
				<input type="text" id="pos-input" list="positions-list" placeholder="Type position and select" />
				<datalist id="positions-list">
					{{range .PositionsList}}
						<option value="{{.}}">
					{{end}}
				</datalist>
				<div id="selected-pos"></div>
				<input type="hidden" name="Positions" id="pos-hidden" value="">
			</div>
			
			<!-- Sort by club checkbox -->
			<label>
				<input type="checkbox" name="sort" id="sort-checkbox" checked /> Sort by club
			</label>
			
			<!-- Only Designated Players checkbox -->
			<label>
				<input type="checkbox" name="dp" id="dp-checkbox" /> Only Designated Players (Compensation ≥ $1,612,500)
			</label>
		</form>
	  </div>
	  <div class="results" id="results">
	    <!-- Filtered results will be injected here via HTMX -->
	  </div>
	</div>

	<script>
	  // Helper function: returns true if the input value exactly matches one of the datalist options.
	  function isValidInput(inputElem, datalistId) {
	      var list = document.getElementById(datalistId);
	      var value = inputElem.value.trim();
	      for (var i = 0; i < list.options.length; i++) {
	          if (list.options[i].value === value) {
	              return true;
	          }
	      }
	      return false;
	  }

	  // Get reference to the form element.
	  var formElem = document.querySelector("form");

	  // --- Auto-complete for Players ---
	  var playerInput = document.getElementById("player-input");
	  var selectedPlayersDiv = document.getElementById("selected-players");
	  var playersHidden = document.getElementById("players-hidden");
	  function updateHiddenPlayers() {
	      var tags = selectedPlayersDiv.querySelectorAll(".tag");
	      var names = [];
	      tags.forEach(function(tag) {
	          names.push(tag.firstChild.textContent.trim());
	      });
	      playersHidden.value = names.join(",");
	      formElem.dispatchEvent(new Event('change'));
	  }
	  playerInput.addEventListener("change", function(e) {
	      var value = playerInput.value.trim();
	      if (value !== "" && isValidInput(playerInput, "players-list")) {
	          var exists = false;
	          selectedPlayersDiv.querySelectorAll(".tag").forEach(function(tag) {
	              if (tag.firstChild.textContent.trim() === value) {
	                  exists = true;
	              }
	          });
	          if (!exists) {
	              var span = document.createElement("span");
	              span.className = "tag";
	              span.textContent = value;
	              var removeBtn = document.createElement("button");
	              removeBtn.type = "button";
	              removeBtn.textContent = "×";
	              removeBtn.addEventListener("click", function() {
	                  span.remove();
	                  updateHiddenPlayers();
	              });
	              span.appendChild(removeBtn);
	              selectedPlayersDiv.appendChild(span);
	              updateHiddenPlayers();
	          }
	      }
	      playerInput.value = "";
	  });

	  // --- Auto-complete for Clubs ---
	  var clubInput = document.getElementById("club-input");
	  var selectedClubsDiv = document.getElementById("selected-clubs");
	  var clubsHidden = document.getElementById("clubs-hidden");
	  function updateHiddenClubs() {
	      var tags = selectedClubsDiv.querySelectorAll(".tag");
	      var names = [];
	      tags.forEach(function(tag) {
	          names.push(tag.firstChild.textContent.trim());
	      });
	      clubsHidden.value = names.join(",");
	      formElem.dispatchEvent(new Event('change'));
	  }
	  clubInput.addEventListener("change", function(e) {
	      var value = clubInput.value.trim();
	      if (value !== "" && isValidInput(clubInput, "clubs-list")) {
	          var exists = false;
	          selectedClubsDiv.querySelectorAll(".tag").forEach(function(tag) {
	              if (tag.firstChild.textContent.trim() === value) {
	                  exists = true;
	              }
	          });
	          if (!exists) {
	              var span = document.createElement("span");
	              span.className = "tag";
	              span.textContent = value;
	              var removeBtn = document.createElement("button");
	              removeBtn.type = "button";
	              removeBtn.textContent = "×";
	              removeBtn.addEventListener("click", function() {
	                  span.remove();
	                  updateHiddenClubs();
	              });
	              span.appendChild(removeBtn);
	              selectedClubsDiv.appendChild(span);
	              updateHiddenClubs();
	          }
	      }
	      clubInput.value = "";
	  });

	  // --- Auto-complete for Positions ---
	  var posInput = document.getElementById("pos-input");
	  var selectedPosDiv = document.getElementById("selected-pos");
	  var posHidden = document.getElementById("pos-hidden");
	  function updateHiddenPos() {
	      var tags = selectedPosDiv.querySelectorAll(".tag");
	      var names = [];
	      tags.forEach(function(tag) {
	          names.push(tag.firstChild.textContent.trim());
	      });
	      posHidden.value = names.join(",");
	      formElem.dispatchEvent(new Event('change'));
	  }
	  posInput.addEventListener("change", function(e) {
	      var value = posInput.value.trim();
	      if (value !== "" && isValidInput(posInput, "positions-list")) {
	          var exists = false;
	          selectedPosDiv.querySelectorAll(".tag").forEach(function(tag) {
	              if (tag.firstChild.textContent.trim() === value) {
	                  exists = true;
	              }
	          });
	          if (!exists) {
	              var span = document.createElement("span");
	              span.className = "tag";
	              span.textContent = value;
	              var removeBtn = document.createElement("button");
	              removeBtn.type = "button";
	              removeBtn.textContent = "×";
	              removeBtn.addEventListener("click", function() {
	                  span.remove();
	                  updateHiddenPos();
	              });
	              span.appendChild(removeBtn);
	              selectedPosDiv.appendChild(span);
	              updateHiddenPos();
	          }
	      }
	      posInput.value = "";
	  });

	  // Trigger an initial change on page load to display results immediately.
	  window.addEventListener("DOMContentLoaded", function() {
	      formElem.dispatchEvent(new Event('change'));
	  });
	</script>
</body>
</html>
`

var resultsHTML = `
<h2>Filtered Players</h2>
<table>
	<thead>
		<tr>
			<th>#</th>
			<th>Club</th>
			<th>Pos</th>
			<th>Name</th>
			<th>Compensation</th>
		</tr>
	</thead>
	<tbody>
		{{ $prevClub := "" }}
		{{ $row := 1 }}
		{{range .Players}}
			{{if and $.Sort (ne $.Sort false) (ne .Club $prevClub)}}
				{{if ne $prevClub ""}}
					<tr><td colspan="5">&nbsp;</td></tr>
					{{ $row = 1 }}
				{{end}}
				{{ $prevClub = .Club }}
			{{end}}
			<tr>
				<td>{{ $row }}</td>
				<td>{{ .Club }}</td>
				<td>{{ .Pos }}</td>
				<td>{{ .Name }}</td>
				<td>{{ .FormattedComp }}</td>
			</tr>
			{{ $row = add $row 1 }}
		{{end}}
	</tbody>
</table>

<h2>Club Totals</h2>
<table>
	<thead>
		<tr>
			<th>#</th>
			<th>Club</th>
			<th>Total Compensation</th>
		</tr>
	</thead>
	<tbody>
		{{range $i, $ct := .ClubTotals}}
		<tr>
			<td>{{add $i 1}}</td>
			<td>{{ $ct.Key }}</td>
			<td>{{commaf $ct.Value}}</td>
		</tr>
		{{end}}
	</tbody>
</table>
`

// =============================================================================
// HTTP Handlers
// =============================================================================

var tmplIndex = template.Must(template.New("index").Funcs(template.FuncMap{
	"eq": func(a, b interface{}) bool { return a == b },
}).Parse(indexHTML))
var tmplResults = template.Must(template.New("results").Funcs(template.FuncMap{
	"commaf": commaf,
	"add":    add,
}).Parse(resultsHTML))

// indexHandler prepares the main page.
// It sorts the data files (newest first), builds DataFileEntry values with formatted display names,
// and computes valid lists for Players, Clubs, and Positions.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	files, err := fs.Glob(dataFS, "data/*_data")
	if err != nil {
		http.Error(w, "Error reading data files", http.StatusInternalServerError)
		return
	}
	var dataFilesEntries []DataFileEntry
	for _, file := range files {
		trimmed := file[len("data/"):]
		dataFilesEntries = append(dataFilesEntries, DataFileEntry{
			Value:   trimmed,
			Display: formatDataFileName(trimmed),
		})
	}
	// Sort data files descending (newest first)
	sort.Slice(dataFilesEntries, func(i, j int) bool {
		return dataFilesEntries[i].Value > dataFilesEntries[j].Value
	})
	// Compute players list from the default (newest) data file
	playersList := []string{}
	if len(dataFilesEntries) > 0 {
		playersData, _, err := processData(dataFilesEntries[0].Value, "", "", "", false, false)
		if err == nil {
			nameSet := make(map[string]struct{})
			for _, p := range playersData {
				nameSet[p.Name] = struct{}{}
			}
			for name := range nameSet {
				playersList = append(playersList, name)
			}
			sort.Strings(playersList)
		}
	}
	// Build clubs list: each club now appears as "Full Name (Abv)"
	var clubsList []string
	for full, abv := range allClubs {
		clubsList = append(clubsList, fmt.Sprintf("%s (%s)", full, abv))
	}
	sort.Strings(clubsList)
	// Positions list: sort the provided list (use full names as provided)
	var positionsList []string = make([]string, len(allPos))
	copy(positionsList, allPos)
	sort.Strings(positionsList)

	data := struct {
		DataFiles     []DataFileEntry
		PlayersList   []string
		ClubsList     []string
		PositionsList []string
	}{
		DataFiles:     dataFilesEntries,
		PlayersList:   playersList,
		ClubsList:     clubsList,
		PositionsList: positionsList,
	}
	if err := tmplIndex.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// playersHandler returns a list of <option> elements for the players datalist,
// reading player names from the specified data file.
func playersHandler(w http.ResponseWriter, r *http.Request) {
	dataFile := r.URL.Query().Get("data")
	if dataFile == "" {
		http.Error(w, "Missing data parameter", http.StatusBadRequest)
		return
	}
	playersData, _, err := processData(dataFile, "", "", "", false, false)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing data: %v", err), http.StatusInternalServerError)
		return
	}
	nameSet := make(map[string]struct{})
	for _, p := range playersData {
		nameSet[p.Name] = struct{}{}
	}
	var names []string
	for name := range nameSet {
		names = append(names, name)
	}
	sort.Strings(names)
	var buf bytes.Buffer
	for _, name := range names {
		buf.WriteString(fmt.Sprintf("<option value=\"%s\">", template.HTMLEscapeString(name)))
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	dataFile := r.FormValue("data")
	clubsStr := r.FormValue("clubs")
	playersStr := r.FormValue("players")
	// Read positions from the hidden field named "Positions"
	posStr := r.FormValue("Positions")
	sortByClub := r.FormValue("sort") != ""
	dp := r.FormValue("dp") != ""

	playersData, clubTotals, err := processData(dataFile, clubsStr, playersStr, posStr, sortByClub, dp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Processing error: %v", err), http.StatusInternalServerError)
		return
	}
	data := struct {
		Players    Players
		ClubTotals []KeyValue
		Sort       bool
	}{
		Players:    playersData,
		ClubTotals: clubTotals.Sort(),
		Sort:       sortByClub,
	}
	if err := tmplResults.Execute(w, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// =============================================================================
// Main
// =============================================================================

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/filter", filterHandler)
	http.HandleFunc("/players", playersHandler)
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
