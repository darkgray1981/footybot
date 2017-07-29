package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

// Google API key/CX for google searches
const GOOGLE_API_KEY = "paste key here"
const GOOGLE_CX = "paste cx here"

// IRC config
const IRC_NICK = "footybot"
const IRC_USER = "botness"
const IRC_SERVER = "irc.synirc.net:6667"
const IRC_CHANNEL = "#epl"

// Main program loop
func main() {
	runBot()
}

// One %d parameter - used to deal with the rate limit
const BASE_BBC_URL = "http://push.api.bbci.co.uk/p?c=%d&t=";

// One %s param - team name
const BASE_TEAM_FIXTURES_URL = "morph://data/bbc-morph-sport-football-scores-tabbed-teams-model/isApp/false/limit/4/team/%s/version/1.0.6"

// One %s param - tournament
const BASE_LEAGUE_FIXTURES_URL = "morph://data/bbc-morph-sport-football-scores-tabbed-model/isApp/false/limit/12/tournament/%s/version/2.0.0"

// Three %s params - endDate, startDate, tournament
const BASE_FIXTURES_URL = "morph://data/bbc-morph-football-scores-match-list-data/endDate/%s/startDate/%s/tournament/%s/version/2.2.1/withPlayerActions/false"

// One %s param - tournament
const BASE_LEAGUE_TABLE_URL = "morph://data/bbc-morph-sport-football-tables-data/competition/%s/version/1.5.0"

// One %s param - team name
const BASE_TEAM_TABLE_URL = "morph://data/bbc-morph-sport-football-table-team-model/team/%s/version/1.0.4"

// BBC API rate limit measures
var bbcRateLimit = 1
var lastApiLookup = time.Now()

// Generate a BBC API URL based on attempts made in the past 30 seconds
func getBbcBaseUrl() string {
	if time.Since(lastApiLookup) < 30 * time.Second {
		bbcRateLimit += 1
	} else {
		bbcRateLimit = 1
	}

	lastApiLookup = time.Now()

	return fmt.Sprintf(BASE_BBC_URL, bbcRateLimit)
}

// Club aliases for ease of use
var aliases = map[string]string{
	"man city":      "manchester city",
	"city":          "manchester city",
	"united":        "manchester united",
	"man utd":       "manchester united",
	"wolves":        "wolverhampton wanderers",
	"wolverhampton": "wolverhampton wanderers",
	"tottenham":     "tottenham hotspur",
	"spurs":         "tottenham hotspur",
	"stoke":         "stoke city",
	"leicester":     "leicester city",
	"qpr":           "queens park rangers",
	"swansea":       "swansea city",
	"west brom":     "west bromwich albion",
	"west ham":      "west ham united",
	"newcastle":     "newcastle united",
	"wigan":         "wigan athletic",
	"charlton":      "charlton athletic",
	"derby":         "derby county",
	"cardiff":       "cardiff city",
	"bolton":        "bolton wanderers",
	"blackburn":     "blackburn rovers",
	"birmingham":    "birmingham city",
	"leeds":         "leeds united",
	"ipswich":       "ipswich town",
	"boro":          "middlesbrough",
	"hull":          "hull city",

	"bournemouth": "afc bournemouth",
	"preston":     "preston north end",

	"farcenal":       "arsenal",
	"farsenal":       "arsenal",
	"gunners":        "arsenal",
	"arse":           "arsenal",
	"arselol":        "arsenal",
	"alwayscheating": "arsenal",
	"absoluteshite":  "stoke city",
	"liverpoo":       "liverpool",
	"manure":         "manchester united",
	"pooshited":      "manchester united",
	"shitty":         "manchester city",
	"chelski":        "chelsea",
	"clowns":         "tottenham hotspur",
	"brizzle":        "bristol city",

	"psg":   "paris st germain",
	"juve":  "juventus",
	"barca": "barcelona",
	"real":  "real-madrid",
}

// URLs for tournaments
var tournaments = map[string]string{
	"UK":    "full-priority-order",
	"AFCON": "africa-cup-of-nations",
	"ARG":   "argentine-primera-division",
	"AUS":   "australian-a-league",
	"AUT":   "austrian-bundesliga",
	"BEL":   "belgian-pro-league",
	"BRA":   "brazilian-league",
	"CL":    "champions-league",
	"CS":    "championship",
	"CC":    "confederations-cup",
	"COPA":  "copa-america",
	"DK":    "danish-superliga",
	"NL":    "dutch-eredivisie",
	"EFLT":  "football-league-trophy",
	"EC":    "european-championship",
	"EU21Q": "euro-under-21-qualifying",
	"EU21":  "euro-under-21-championship",
	"EL":    "europa-league",
	"ECQ":   "european-championship-qualifying",
	"FIN":   "finnish-veikkausliiga",
	"FR":    "french-ligue-one",
	"DE":    "german-bundesliga",
	"GOLD":  "gold-cup",
	"GRC":   "greek-superleague",
	"HIGH":  "highland-league",
	"LOI":   "league-of-ireland-premier",
	"IP":    "irish-premiership",
	"L1":    "league-one",
	"L2":    "league-two",
	"LOW":   "lowland-league",
	"MOLY":  "olympic-football-men",
	"CONF":  "national-league",
	"CONFN": "national-league-north",
	"CONFS": "national-league-south",
	"NO":    "norwegian-tippeligaen",
	"PT":    "portuguese-primeira-liga",
	"PL":    "premier-league",
	"EPL":   "premier-league",
	"EN":    "premier-league",
	"RPL":   "russian-premier-league",
	"SCLC":  "scottish-league-cup",
	"SL1":   "scottish-league-one",
	"SL2":   "scottish-league-two",
	"SPL":   "scottish-premiership",
	"SHE":   "shebelieves-cup",
	"ES":    "spanish-la-liga",
	"SE":    "swedish-allsvenskan",
	"SW":    "swiss-super-league",
	"TK":    "turkish-super-lig",
	"US":    "us-major-league",
	"MLS":   "us-major-league",
	"WPL":   "welsh-premier-league",
	"WEC":   "womens-european-championship",
	"WECQ":  "womens-european-championship-qualifying",
	"WOLY":  "olympic-football-women",
	"WPN":   "womens-premier-league-north",
	"WPS":   "womens-premier-league-south",
	"WSL1":  "womens-super-league",
	"WSL2":  "womens-super-league-two",
	"WWCQ":  "womens-world-cup-qualifying-european",
	"WWC":   "womens-world-cup",
	"WC":    "world-cup",
	"WCQE":  "world-cup-qualifying-european",
}

// Check if an alias is registered, else strip and return
func checkAlias(team string) string {
	team = strings.TrimSpace(strings.ToLower(team))

	if team == "scum" {
		// Grab random club from the alias list
		for _, club := range aliases {
			team = club
			break
		}
	} else if club, ok := aliases[team]; ok {
		team = club
	}

	team = strings.Replace(team, " ", "-", -1)

	return team
}

// Aliases for mapping currency names
var currencies = map[string]string{
	"yen":      "JPY",
	"dollar":   "USD",
	"dollars":  "USD",
	"bucks":    "USD",
	"bux":      "USD",
	"euro":     "EUR",
	"euros":    "EUR",
	"crowns":   "SEK",
	"pound":    "GBP",
	"pounds":   "GBP",
	"quid":     "GBP",
	"bitcoin":  "BTC",
	"bitcoins": "BTC",
}

// Check if an alias is registered, else strip and return
func checkCurrency(name string) string {
	name = strings.ToLower(name)

	if acronym, ok := currencies[name]; ok {
		name = acronym
	}

	name = strings.ToUpper(name)

	return name
}

// Format float in a human way
func humanize(f float64) string {

	sign := ""
	if f < 0 {
		sign = "-"
		f = -f
	}

	n := uint64(f)

	// Grab two rounded decimals
	decimals := uint64((f+0.005)*100) % 100

	var buf []byte

	if n == 0 {
		buf = []byte{'0'}
	} else {
		buf = make([]byte, 0, 16)

		for n >= 1000 {
			for i := 0; i < 3; i++ {
				buf = append(buf, byte(n%10)+'0')
				n /= 10
			}

			buf = append(buf, ',')
		}

		for n > 0 {
			buf = append(buf, byte(n%10)+'0')
			n /= 10
		}
	}

	// Reverse the byte slice
	for l, r := 0, len(buf)-1; l < r; l, r = l+1, r-1 {
		buf[l], buf[r] = buf[r], buf[l]
	}

	return fmt.Sprintf("%s%s.%02d", sign, buf, decimals)
}

// Return Yahoo currency conversion
func Currency(query string) string {
	yahoo := "http://download.finance.yahoo.com/d/quotes.csv?f=l1&e=.csv&s="

	parts := strings.Split(strings.TrimSpace(query), " ")
	if len(parts) != 4 {
		return "Error - Malformed query (ex. 100 JPY in USD)"
	}

	r := strings.NewReplacer(",", "", "K", "e3", "M", "e6", "B", "e9")

	multiplier, err := strconv.ParseFloat(r.Replace(strings.ToUpper(strings.TrimSpace(parts[0]))), 64)
	if err != nil {
		return "Error - " + err.Error()
	}

	from := checkCurrency(parts[1])
	to := checkCurrency(parts[3])

	queryUrl := yahoo + from + to + "=X"

	resp, err := http.Get(queryUrl)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	if resp.StatusCode != 200 {
		return "Error - Something went wrong"
		fmt.Println("Yahoo error dump: ", string(data))
	}

	number, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		if strings.TrimSpace(string(data)) == "N/A" {
			return "Error - Unknown currency"
		}
		return "Error - " + err.Error()
	}

	return fmt.Sprintf("%s %s is \x02%s\x02 %s", parts[0], from, humanize(multiplier*number), to)
}

// Return latest results for a club
func LatestResults(team string) string {
	var matches *footballMatches
	matches, errorMsg := ParseBbcFixtures(team)
	if matches == nil {
		return errorMsg
	}

	if len(matches.results) == 0 {
		return "Error - no fixtures for " + team
	}

	resultsArr := make([]string, 0)

	for i := 0; i < len(matches.results); i++ {
		resultsArr = append(resultsArr, formatMatchResult(matches.results[i]))
	}

	return strings.Join(resultsArr[:], ", ")
}

// Return current league table position for a club
func TablePosition(team string) string {
	team = checkAlias(team)

	var site = getBbcBaseUrl() + url.QueryEscape(fmt.Sprintf(BASE_TEAM_TABLE_URL, team))
	resp, err := http.Get(site)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "Error - Table not found."
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	pushResponse := bbcPushResponse{}
	json.Unmarshal(body, &pushResponse)
	if len(pushResponse.Moments) == 0 || len(pushResponse.Moments[0].Payload) == 0 {
		fmt.Println(site + " " + string(body))
		return "No JSON table payload, team table not found"
	}

	leagueTable := teamLeagueTable{}
	json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &leagueTable)

	for i := 0; i < len(leagueTable); i++ {
		for j := 0; j < len(leagueTable[i].Tables); j++ {
			teams := leagueTable[i].Tables[j].Teams
			for k := 0; k < len(teams); k++ {
				if teams[k].Slug == team {
					compName := leagueTable[i].Tournament.Name.Abbreviation
					teamName := teams[k].Name.Abbreviation
					position := teams[k].Rank.Current
					played := teams[k].Stats.Played
					won := teams[k].Stats.Won
					drawn := teams[k].Stats.Drawn
					lost := teams[k].Stats.Lost
					goalsFor := teams[k].Stats.GoalsFor
					goalsAgainst := teams[k].Stats.GoalsAgainst
					goalDiff := teams[k].Stats.GoalDifference
					points := teams[k].Stats.Points
					return fmt.Sprintf("%s #%d \x02%s\x02 - P: %d, W: %d, D: %d, L: %d, F: %d, A: %d, GD: %d, Pts: %d", compName, position, teamName, played, won, drawn, lost, goalsFor, goalsAgainst, goalDiff, points)
				}
			}
		}
	}

	return "Team table not found"
}

// Parser for the BBC API fixtures response
func ParseBbcFixtures(team string) (*footballMatches, string) {
	parsedResponse := &footballMatches{}

	var isTournament = false
	if tournament, ok := tournaments[strings.ToUpper(team)]; ok {
		team = tournament
		isTournament = true
	} else {
		team = checkAlias(team)
	}

	var site = ""
	if isTournament {
		site = getBbcBaseUrl() + url.QueryEscape(fmt.Sprintf(BASE_LEAGUE_FIXTURES_URL, team))
	} else {
		site = getBbcBaseUrl() + url.QueryEscape(fmt.Sprintf(BASE_TEAM_FIXTURES_URL, team))
	}

	resp, err := http.Get(site)
	if err != nil {
		return nil, "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Club not found: " + team)
		return nil, "Error - Club not found."
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "Error - " + err.Error()
	}

	pushResponse := bbcPushResponse{}
	json.Unmarshal(body, &pushResponse)
	if len(pushResponse.Moments) == 0 || len(pushResponse.Moments[0].Payload) == 0 {
		fmt.Println(site + " " + string(body))
		return nil, "No JSON fixture payload, no fixtures found"
	}

	// Tournament & Team responses differ in wrapper structure
	if isTournament {
		payload := tournamentMatches{}
		json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &payload)

		for i := 0; i < len(payload.Fixtures.Tournament.Stages); i++ {
			stages := payload.Fixtures.Tournament.Stages[i].Rounds
			for j := 0; j < len(stages); j++ {
				matches := stages[j].Events
				for k := 0; k < len(matches); k++ {
					match := footballMatch{}
					match.kickOffTime = matches[k].StartTime
					match.isTournamentGame = true
					match.Tournament = payload.Fixtures.Tournament.Name.Abbreviation
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
					match.isFixture = matches[k].EventProgress.Status == "FIXTURE"
					match.inProgress = matches[k].EventProgress.Status == "LIVE"
					match.isResult = matches[k].EventProgress.Status == "RESULT"
					parsedResponse.fixtures = append(parsedResponse.fixtures, match)
				}
			}
		}

		for i := 0; i < len(payload.Results.Tournament.Stages); i++ {
			stages := payload.Results.Tournament.Stages[i].Rounds
			for j := 0; j < len(stages); j++ {
				matches := stages[j].Events
				for k := 0; k < len(matches); k++ {
					match := footballMatch{}
					match.kickOffTime = matches[k].StartTime
					match.isTournamentGame = true
					match.Tournament = payload.Results.Tournament.Name.Abbreviation
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
					match.isFixture = matches[k].EventProgress.Status == "FIXTURE"
					match.inProgress = matches[k].EventProgress.Status == "LIVE"
					match.isResult = matches[k].EventProgress.Status == "RESULT"
					parsedResponse.results = append(parsedResponse.results, match)
				}
			}
		}

		var inProgress []footballMatch
		for i := 0; i < len(payload.Today.Tournament.Stages); i++ {
			stages := payload.Today.Tournament.Stages[i].Rounds
			for j := 0; j < len(stages); j++ {
				matches := stages[j].Events
				for k := 0; k < len(matches); k++ {
					match := footballMatch{}
					match.kickOffTime = matches[k].StartTime
					match.isTournamentGame = true
					match.Tournament = payload.Results.Tournament.Name.Abbreviation
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
					match.isFixture = matches[k].EventProgress.Status == "FIXTURE"
					match.inProgress = matches[k].EventProgress.Status == "LIVE"
					match.isResult = matches[k].EventProgress.Status == "RESULT"
					match.minutesElapsed = matches[k].MinutesElapsed

					if match.inProgress {
						inProgress = append(inProgress, match)
					} else {
						parsedResponse.results = append([]footballMatch{match}, parsedResponse.results...)
					}
				}
			}
		}

		parsedResponse.fixtures = append(inProgress, parsedResponse.fixtures...)
	} else {
		payload := teamMatches{}
		json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &payload)

		for i := 0; i < len(payload.Fixtures.Body.Rounds); i++ {
			matches := payload.Fixtures.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Fixtures.Body.Rounds[i].Name.Abbreviation
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				match.isFixture = matches[j].EventProgress.Status == "FIXTURE"
				match.inProgress = matches[j].EventProgress.Status == "LIVE"
				match.isResult = matches[j].EventProgress.Status == "RESULT"
				parsedResponse.fixtures = append(parsedResponse.fixtures, match)
			}
		}

		for i := 0; i < len(payload.Results.Body.Rounds); i++ {
			matches := payload.Results.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Results.Body.Rounds[i].Name.Abbreviation
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				match.isFixture = matches[j].EventProgress.Status == "FIXTURE"
				match.inProgress = matches[j].EventProgress.Status == "LIVE"
				match.isResult = matches[j].EventProgress.Status == "RESULT"
				parsedResponse.results = append(parsedResponse.results, match)
			}
		}

		for i := 0; i < len(payload.Today.Body.Rounds); i++ {
			matches := payload.Today.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Results.Body.Rounds[i].Name.Abbreviation
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				match.isFixture = matches[j].EventProgress.Status == "FIXTURE"
				match.inProgress = matches[j].EventProgress.Status == "LIVE"
				match.isResult = matches[j].EventProgress.Status == "RESULT"
				match.minutesElapsed = matches[j].MinutesElapsed

				if match.inProgress {
					parsedResponse.fixtures = append([]footballMatch{match}, parsedResponse.fixtures...)
				} else {
					parsedResponse.results = append([]footballMatch{match}, parsedResponse.results...)
				}
			}
		}
	}

	return parsedResponse, ""
}

// Get next matches on a club's schedule
func NextMatch(team string) string {
	var matches *footballMatches
	matches, errorMsg := ParseBbcFixtures(team)
	if matches == nil {
		return errorMsg
	}

	if len(matches.fixtures) == 0 {
		return "Error - no fixtures for " + team
	}

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return "Error - " + err.Error()
	}

	fixturesArr := make([]string, 0)

	for i := 0; i < len(matches.fixtures); i++ {
		match := matches.fixtures[i]

		var compName = ""
		if !match.isTournamentGame {
			compName = " - " + match.Tournament
		}

		if match.inProgress {
			fixturesArr = append(fixturesArr, formatMatchResult(match))
		} else {
			var kickOffTime = match.kickOffTime.In(loc).Format("15:04 Jan 2")
			var homeTeam = match.HomeTeam.Name.First
			var awayTeam = match.AwayTeam.Name.First

			fixturesArr = append(fixturesArr, fmt.Sprintf("\x02%s\x02 vs \x02%s\x02 (%s%s)", homeTeam, awayTeam, kickOffTime, compName))
		}
	}

	return strings.Join(fixturesArr[:], ", ")
}

func formatMatchResult(match footballMatch) string {
	var matchResult = ""

	var compName = ""
	if !match.isTournamentGame {
		compName = " (" + match.Tournament + ")"
	}

	var homeTeam = match.HomeTeam.Name.First
	var homeGoals = match.HomeTeam.Scores.Score
	var awayTeam = match.AwayTeam.Name.First
	var awayGoals = match.AwayTeam.Scores.Score
	var progressStr = "FT"
	if match.inProgress {
		progressStr = fmt.Sprintf("'%d", match.minutesElapsed)
	}

	if match.HomeTeam.Scores.ExtraTime != nil && match.AwayTeam.Scores.ExtraTime != nil {
		var homeGoalsEt = match.HomeTeam.Scores.ExtraTime.(float64)
		var awayGoalsEt = match.AwayTeam.Scores.ExtraTime.(float64)
		if match.HomeTeam.Scores.Shootout != nil && match.AwayTeam.Scores.Shootout != nil {
			var homeGoalsShootout = match.HomeTeam.Scores.Shootout.(float64)
			var awayGoalsShootout = match.AwayTeam.Scores.Shootout.(float64)
			if match.inProgress {
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 (Pens %d-%d)%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), awayTeam, int(homeGoalsShootout), int(awayGoalsShootout), compName)
			} else {
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 AET (Pens %d-%d)%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), awayTeam, int(homeGoalsShootout), int(awayGoalsShootout), compName)
			}
		} else {
			if match.inProgress {
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 %s%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), awayTeam, progressStr, compName)
			} else {
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 AET%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), awayTeam, compName)
			}
		}
	} else {
		matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 %s%s", homeTeam, homeGoals, awayGoals, awayTeam, progressStr, compName)
	}

	return matchResult
}

// Figure out current time in the UK
func GetUKTime() string {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return "Error - " + err.Error()
	}

	return time.Now().In(loc).Format("15:04:05 MST")
}

// Generate fixtures offset date in UK locale
func getUKDate(offset int) string {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return "Error - " + err.Error()
	}

	return time.Now().In(loc).AddDate(0, 0, offset*1).Format("2006-01-02")
}

// Run a Google search
func Google(query string) string {

	type GoogleResult struct {
		Items []struct {
			Title   string
			Link    string
			Snippet string
		}
	}

	// Glue the query together
	queryUrl := "https://www.googleapis.com/customsearch/v1?key=" + GOOGLE_API_KEY + "&cx=" + GOOGLE_CX
	queryUrl += "&q=" + url.QueryEscape(query)
	queryUrl += "&fields=items(title,link,snippet)&safe=off&num=1"

	resp, err := http.Get(queryUrl)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	if resp.StatusCode != 200 {
		return "Error - Something went wrong (quota exceeded?)"
		fmt.Println("Google error dump: ", string(data))
	}

	var result GoogleResult

	err = json.Unmarshal(data, &result)
	if err != nil {
		return "Error - " + err.Error()
	}

	if len(result.Items) == 0 {
		return "Error - No results"
	}

	return fmt.Sprintf("%s -- \x02%s\x02: \"%s\"", result.Items[0].Link, result.Items[0].Title, strings.Replace(result.Items[0].Snippet, "\n", "", -1))
}

// Return current league table positions
func ShowTable(args string) string {

	maxLen := 2
	if strings.HasPrefix(args, "#") {
		maxLen = 3
	}
	zone := "EN"  // Default to Premier League
	subZone := "" // No default subzone
	pos := "1-6"  // Default to show pos 1-6
	splitArgs := strings.SplitN(args, " ", maxLen)

	if len(splitArgs) < 1 {
		return "Error - No parameters"
	} else if !strings.HasPrefix(args, "#") && tournaments[strings.ToUpper(splitArgs[0])] == "" {
		return TablePosition(args)
	} else {
		if strings.HasPrefix(splitArgs[0], "#") {
			pos = splitArgs[0][1:]

			if len(splitArgs) > 1 {
				zone = strings.ToUpper(splitArgs[1])
				if len(splitArgs) > 2 {
					subZone = strings.ToUpper(splitArgs[2])
				}
			}
		} else {
			if splitArgs[0] != "" {
				zone = strings.ToUpper(splitArgs[0])
				if len(splitArgs) > 1 {
					subZone = strings.ToUpper(splitArgs[1])
				}
			}
		}
	}

	positions := strings.Split(pos, "-")
	var fromPos = 1
	var toPos = 6
	if len(positions) == 2 && len(positions[1]) > 0 {
		if from, err := strconv.Atoi(positions[0]); err == nil {
			fromPos = from
		}
		if to, err := strconv.Atoi(positions[1]); err == nil {
			toPos = to
		}
	} else if len(positions) == 1 {
		if from, err := strconv.Atoi(positions[0]); err == nil {
			fromPos = from
			toPos = fromPos
		}
	}

	if fromPos < 1 || toPos < 1 || fromPos > toPos {
		return "Invalid table positions"
	}

	if _, ok := tournaments[zone]; !ok {
		return "Error - Unknown zone"
	}

	var site = getBbcBaseUrl() + url.QueryEscape(fmt.Sprintf(BASE_LEAGUE_TABLE_URL, tournaments[zone]))
	resp, err := http.Get(site)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "Error - Table not found."
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	pushResponse := bbcPushResponse{}
	json.Unmarshal(body, &pushResponse)
	if len(pushResponse.Moments) == 0 || len(pushResponse.Moments[0].Payload) == 0 {
		fmt.Println(site + " " + string(body))
		return "No JSON table payload, table not found"
	}

	table := leagueTable{}
	json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &table)

	tableStr := make([]string, 0)
	var multiTable = len(table.SportTables.Tables) > 1
	tableIndex := 0

	for i := 0; i < len(table.SportTables.Tables); i++ {
		if multiTable && subZone != "" {
			if strings.HasPrefix(strings.ToUpper(table.SportTables.Tables[i].Group.Name), subZone) {
				tableIndex = i
			} else {
				continue
			}
		}

		var rows = table.SportTables.Tables[tableIndex].Rows
		if toPos > len(rows) {
			return "Invalid table positions"
		}

		for j := fromPos; j <= toPos; j++ {
			var team = rows[j-1];
			var teamName = team.Cells[2].Td.AbbrLink.Abbr;
			if teamName == "" {
				teamName = team.Cells[2].Td.Abbr
			}

			tableStr = append(tableStr, fmt.Sprintf("#%d \x02%s\x02 - P: %d, GD: %d, Pts: %d", j, teamName, team.Cells[3].Td.Text, team.Cells[9].Td.Text, team.Cells[10].Td.Text))
		}

		break
	}

	return strings.Join(tableStr, " | ")
}

// Get upcoming games from all leagues
func AllFixtures(zone, input string) string {

	input = strings.TrimSpace(input)
	wantedDate := 0
	if len(input) > 0 {
		var err error
		wantedDate, err = strconv.Atoi(input)
		if err != nil {
			return "Error - " + err.Error()
		}
	}

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return "Error - " + err.Error()
	}

	dateStr := getUKDate(wantedDate)

	site := getBbcBaseUrl() + url.QueryEscape(fmt.Sprintf(BASE_FIXTURES_URL, dateStr, dateStr, tournaments[zone]))

	resp, err := http.Get(site)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "Error - Fixtures not found."
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	pushResponse := bbcPushResponse{}
	json.Unmarshal(body, &pushResponse)
	if len(pushResponse.Moments) == 0 || len(pushResponse.Moments[0].Payload) == 0 {
		fmt.Println(site)
		fmt.Println(string(body))
		return "No JSON fixtures payload, fixtures not found"
	}

	payload := fixtureList{}
	json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &payload)

	var fixtures []footballMatch
	for i := 0; i < len(payload.MatchData); i++ {
		for days := range payload.MatchData[i].TournamentDatesWithEvents {
			daysGames := payload.MatchData[i].TournamentDatesWithEvents[days]
			for j := 0; j < len(daysGames); j++ {
				for k := 0; k < len(daysGames[j].Events); k++ {
					match := footballMatch{}
					match.kickOffTime = daysGames[j].Events[k].StartTime
					match.Tournament = payload.MatchData[i].TournamentMeta.TournamentName.Abbreviation
					match.HomeTeam = daysGames[j].Events[k].HomeTeam
					match.AwayTeam = daysGames[j].Events[k].AwayTeam
					match.isFixture = daysGames[j].Events[k].EventProgress.Status == "FIXTURE"
					match.inProgress = daysGames[j].Events[k].EventProgress.Status == "LIVE"
					match.isResult = daysGames[j].Events[k].EventProgress.Status == "RESULT"
					fixtures = append(fixtures, match)
				}
			}
		}
	}

	matchesArr := make([]string, 0)

	for i := 0; i < len(fixtures); i++ {
		if fixtures[i].isResult || fixtures[i].inProgress {
			matchesArr = append(matchesArr, formatMatchResult(fixtures[i]))
		} else {
			var kickOffTime = fixtures[i].kickOffTime.In(loc).Format("15:04 Jan 2")
			var homeTeam = fixtures[i].HomeTeam.Name.First
			var awayTeam = fixtures[i].AwayTeam.Name.First
			var compName = ""
			if zone == "UK" {
				compName = " " + fixtures[i].Tournament
			}

			matchesArr = append(matchesArr, fmt.Sprintf("\x02%s\x02 vs \x02%s\x02 (%s%s)", homeTeam, awayTeam, kickOffTime, compName))
		}
	}

	return strings.Join(matchesArr[:], ", ")
}

// Run the bot
func runBot() {
	// Set nick and username
	bot := irc.IRC(IRC_NICK, IRC_USER)
	bot.Debug = false

	err := bot.Connect(IRC_SERVER)
	if err != nil {
		log.Fatal("Can't connect to IRC server!", err.Error())
	}

	// Register channel join action on connect
	bot.AddCallback("001", func(e *irc.Event) { bot.Join(IRC_CHANNEL) })

	// Handle next three games query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".next "
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := NextMatch(event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle table position query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".table "
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		arg := event.Message()[len(cmd):]
		var result = ShowTable(arg)

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle latest results query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".results "
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := LatestResults(event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle time query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".time"
		if event.Message() != cmd {
			return
		}

		result := GetUKTime()

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle google query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".g "
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := Google(event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle currency query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".c "
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := Currency(event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all UK fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".uk"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("UK", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all UK fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".epl"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("PL", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all CL fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".cl"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("CL", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all EL fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".el"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("EL", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all La Liga fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".es"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("ES", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all Serie A fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".it"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("IT", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all Serie A fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".us"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("US", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all Serie A fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".fr"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("FR", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all Serie A fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".de"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("DE", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all Serie A fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".nl"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("NL", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all A League fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".au"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("AUS", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle all European Championshio fixtures query
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".ec"
		if !strings.HasPrefix(event.Message(), cmd) {
			return
		}

		result := AllFixtures("EC", event.Message()[len(cmd):])

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	// Handle help list
	bot.AddCallback("PRIVMSG", func(event *irc.Event) {
		cmd := ".help"
		if event.Message() != cmd {
			return
		}

		result := ".au .cl .de .ec .el .es .fr .it .next .nl .results .table .time .uk .us .c (currency) .g (google)"

		message := event.Nick + ": " + result

		bot.Privmsg(event.Arguments[0], message)
	})

	bot.Loop()
}
