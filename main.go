package main

import (
	"bytes"
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
	"golang.org/x/net/html"
)

// Google API key/CX for google searches
const GOOGLE_API_KEY = "paste key here"
const GOOGLE_CX = "paste cx here"

// IRC config
const IRC_NICK = "footybot-JB"
const IRC_USER = "botness"
const IRC_SERVER = "irc.synirc.net:6667"
const IRC_CHANNEL = "#epl-test"

// Main program loop
func main() {
	runBot()
}

// Note - possibly rate limited
const BASE_BBC_URL = "http://push.api.bbci.co.uk/p?c=1&t=";

// One %s param - team name
const BASE_TEAM_FIXTURES_URL = "morph://data/bbc-morph-sport-football-scores-tabbed-teams-model/isApp/false/limit/4/team/%s/version/1.0.6"

// One %s param - tournament
const BASE_LEAGUE_FIXTURES_URL = "morph://data/bbc-morph-sport-football-scores-tabbed-model/isApp/false/limit/12/tournament/%s/version/2.0.0"

// Three %s params - endDate, startDate, tournament
const BASE_FIXTURES_URL = "morph://data/bbc-morph-football-scores-match-list-data/endDate/%s/startDate/%s/tournament/%s/version/2.2.1/withPlayerActions/false"

// Two %s params - endDate startDate
const BASE_ALL_FIXTURES_URL = "morph://data/bbc-morph-football-scores-match-list-data/endDate/%s/startDate/%s/tournament/full-priority-order/version/2.2.1/withPlayerActions/false"

// One %s param = competition
const BASE_TABLE_URL = "morph://data/bbc-morph-sport-football-tables-data/competition/%s/version/1.4.1"

var aliases = map[string]string {
	"city":          "manchester city",
	"united":        "manchester united",
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

	"bournemouth":   "afc bournemouth",
	"preston":       "preston north end",

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

	"psg": "paris st germain",
	"juve": "juventus",
	"barca": "barcelona",
	"real": "real-madrid",
}

// URLs for tournaments
var tournaments = map[string]string{
	"AFCON":    "africa-cup-of-nations",
	"ARG":      "argentine-primera-division",
	"AUS":      "australian-a-league",
	"AUT":      "austrian-bundesliga",
	"BEL":      "belgian-pro-league",
	"BRA":      "brazilian-league",
	"CL":       "champions-league",
	"CS":       "championship",
	"CC":       "confederations-cup",
	"COPA":     "copa-america",
	"DK":       "danish-superliga",
	"NL":       "dutch-eredivisie",
	"EFLT":     "football-league-trophy",
	"EC":       "european-championship",
	"EU21Q":    "euro-under-21-qualifying",
	"EU21":     "euro-under-21-championship",
	"EL":       "europa-league",
	"ECQ":      "european-championship-qualifying",
	"FIN":      "finnish-veikkausliiga",
	"FR":       "french-ligue-one",
	"DE":       "german-bundesliga",
	"GOLD":     "gold-cup",
	"GRC":      "greek-superleague",
	"HIGH":     "highland-league",
	"LOI":      "league-of-ireland-premier",
	"IP":       "irish-premiership",
	"L1":       "league-one",
	"L2":       "league-two",
	"LOW":      "lowland-league",
	"MOLY":     "olympic-football-men",
	"CONF":     "national-league",
	"CONFN":    "national-league-north",
	"CONFS":    "national-league-south",
	"NO":       "norwegian-tippeligaen",
	"PT":       "portuguese-primeira-liga",
	"PL":       "premier-league",
	"EN":       "premier-league",
	"RPL":      "russian-premier-league",
	"SCLC":     "scottish-league-cup",
	"SL1":      "scottish-league-one",
	"SL2":      "scottish-league-two",
	"SPL":      "scottish-premiership",
	"SHE":      "shebelieves-cup",
	"ES":       "spanish-la-liga",
	"SE":       "swedish-allsvenskan",
	"SW":       "swiss-super-league",
	"TK":       "turkish-super-lig",
	"US":       "us-major-league",
	"MLS":       "us-major-league",
	"WPL":      "welsh-premier-league",
	"WEC":      "womens-european-championship",
	"WECQ":     "womens-european-championship-qualifying",
	"WOLY":     "olympic-football-women",
	"WPN":      "womens-premier-league-north",
	"WPS":      "womens-premier-league-south",
	"WSL1":     "womens-super-league",
	"WSL2":     "womens-super-league-two",
	"WWCQ":     "womens-world-cup-qualifying-european",
	"WWC":      "womens-world-cup",
	"WC":       "world-cup",
	"WCQE":     "world-cup-qualifying-european",
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

	for i := 0; i<len(matches.results); i++ {
		resultsArr = append(resultsArr, formatMatchResult(matches.results[i]))
	}

	return strings.Join(resultsArr[:], ", ")
}

// Return current league table position for a club
func TablePosition(team string) string {
	team = checkAlias(team)

	resp, err := http.Get("http://www.bbc.co.uk/sport/football/teams/" + team + "/fixtures")
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Club not found: " + team)
		return "Error - Club not found."
	}

	// Run HTML parser to get DOM
	root, err := html.Parse(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	// Map to store details of interest
	m := make(map[string]string)

	// Helper function to save text nodes in a subtree
	var pr func(string, *html.Node)
	pr = func(parentClass string, n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				trimmed := strings.TrimSpace(c.Data)
				if len(trimmed) > 0 {
					m[parentClass] = trimmed
				}
			}

			// Retain parent's class name unless we have one of our own
			childClass := parentClass
			for _, a := range c.Attr {
				if a.Key == "class" {
					childClass = a.Val
					break
				}
			}

			pr(childClass, c)
		}
	}

	// Helper function to look through DOM for elements of interest
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			// Grab club details
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "current-team") {
					pr(a.Val, n)
					break
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "select" {
			// Grab league name
			for _, a := range n.Attr {
				if a.Key == "id" && a.Val == "competitionFilter" {
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						for _, attr := range c.Attr {
							if attr.Key == "selected" {
								pr(a.Val, c)
								break
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	if len(m) == 0 {
		fmt.Println("Table not found: " + team)
		return "Error - No table found."
	}

	return fmt.Sprintf("%s #%s \x02%s\x02 - P: %s, GD: %s, Pts: %s", m["competitionFilter"], m["position-number"], m["team-name"], m["played"], m["goal-difference"], m["points"])
}

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
		site = BASE_BBC_URL + url.QueryEscape(fmt.Sprintf(BASE_LEAGUE_FIXTURES_URL, team))
	} else {
		site = BASE_BBC_URL + url.QueryEscape(fmt.Sprintf(BASE_TEAM_FIXTURES_URL, team))
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
		fmt.Println(string(body))
		return nil, "Error - failed to parse BBC JSON response"
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
					match.Tournament = payload.Fixtures.Tournament.Name.First
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
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
					match.Tournament = payload.Results.Tournament.Name.First
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
					parsedResponse.results = append(parsedResponse.results, match)
				}
			}
		}

		for i := 0; i < len(payload.Today.Tournament.Stages); i++ {
			stages := payload.Today.Tournament.Stages[i].Rounds
			for j := 0; j < len(stages); j++ {
				matches := stages[j].Events
				for k := 0; k < len(matches); k++ {
					match := footballMatch{}
					match.kickOffTime = matches[k].StartTime
					match.isTournamentGame = true
					match.Tournament = payload.Results.Tournament.Name.First
					match.HomeTeam = matches[k].HomeTeam
					match.AwayTeam = matches[k].AwayTeam
					match.inProgress = matches[k].EventProgress.Status != "RESULT"
					match.minutesElapsed = matches[k].MinutesElapsed

					if match.inProgress {
						parsedResponse.fixtures = append([]footballMatch{match}, parsedResponse.fixtures...)
					} else {
						parsedResponse.results = append([]footballMatch{match}, parsedResponse.results...)
					}
				}
			}
		}
	} else {
		payload := teamMatches{}
		json.Unmarshal([]byte(pushResponse.Moments[0].Payload), &payload)

		for i := 0; i < len(payload.Fixtures.Body.Rounds); i++ {
			matches := payload.Fixtures.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Fixtures.Body.Rounds[i].Name.First
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				parsedResponse.fixtures = append(parsedResponse.fixtures, match)
			}
		}

		for i := 0; i < len(payload.Results.Body.Rounds); i++ {
			matches := payload.Results.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Results.Body.Rounds[i].Name.First
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				parsedResponse.results = append(parsedResponse.results, match)
			}
		}

		for i := 0; i < len(payload.Today.Body.Rounds); i++ {
			matches := payload.Today.Body.Rounds[i].Events
			for j := 0; j < len(matches); j++ {
				match := footballMatch{}
				match.kickOffTime = matches[j].StartTime
				match.Tournament = payload.Results.Body.Rounds[i].Name.First
				match.HomeTeam = matches[j].HomeTeam
				match.AwayTeam = matches[j].AwayTeam
				match.inProgress = matches[j].EventProgress.Status != "RESULT"
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

	for i := 0; i<len(matches.fixtures); i++ {
		match := matches.fixtures[i]

		var compName = ""
		if !match.isTournamentGame {
			compName = " - " + match.Tournament
		}

		if match.inProgress {
			fixturesArr = append(fixturesArr, formatMatchResult(match))
		} else {
			var kickOffTime = match.kickOffTime.In(loc).Format("15:04 Aug 2")
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
	var progressStr = ""
	if match.inProgress {
		progressStr = fmt.Sprintf(" '%d", match.minutesElapsed)
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
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02%s%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), progressStr, awayTeam, compName)
			} else {
				matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02 AET%s", homeTeam, int(homeGoalsEt), int(awayGoalsEt), awayTeam, compName)
			}
		}
	} else {
		matchResult = fmt.Sprintf("\x02%s\x02 %d-%d \x02%s\x02%s%s", homeTeam, homeGoals, awayGoals, awayTeam, progressStr, compName)
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

	zone := "EN" // Default to Premier League
	var pos string
	splitargs := strings.SplitN(args, " ", 2)

	if len(splitargs) < 1 {
		return "Error - No parameters"
	} else {
		pos = splitargs[0]

		if len(splitargs) == 2 {
			zone = strings.ToUpper(splitargs[1])
		}
	}

	if _, ok := tournaments[zone]; !ok {
		return "Error - Unknown zone"
	}

	resp, err := http.Get(tournaments[zone])
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("League not found")
		return "Error - League not found."
	}

	// Run HTML parser to get DOM
	root, err := html.Parse(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	// Map to store entire league
	league := make(map[string]map[string]string)

	// Helper function to save text nodes in a subtree
	var pr func(string, *html.Node, map[string]string)
	pr = func(parentClass string, n *html.Node, m map[string]string) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				trimmed := strings.TrimSpace(c.Data)
				if len(trimmed) > 0 {
					m[parentClass] = trimmed
				}
			}

			// Retain parent's class name unless we have one of our own
			childClass := parentClass
			for _, a := range c.Attr {
				if a.Key == "class" {
					childClass = a.Val
					break
				}
			}

			pr(childClass, c, m)
		}
	}

	// Helper function to look through DOM for elements of interest
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			// Grab club details
			for _, a := range n.Attr {
				if a.Key == "class" && strings.HasPrefix(a.Val, "team") {
					m := make(map[string]string)
					pr(a.Val, n, m)
					if len(m) != 0 {
						if _, ok := league[m["position-number"]]; !ok {
							league[m["position-number"]] = m
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	if len(league) == 0 {
		fmt.Println("League table not found")
		return "Error - No league table found."
	}

	result := "Error - position not found."

	positions := strings.Split(pos, "-")
	if len(positions) == 2 && len(positions[1]) > 0 {
		var buf bytes.Buffer

		from, _ := strconv.Atoi(positions[0])
		to, _ := strconv.Atoi(positions[1])

		for i := from; i <= to; i++ {
			if m, ok := league[strconv.Itoa(i)]; ok {
				if i > from {
					buf.WriteString(" | ")
				}

				s := fmt.Sprintf("#%s \x02%s\x02 (%sp)", m["position-number"], m["team-name"], m["points"])
				buf.WriteString(s)

			} else {
				return result
			}
		}

		result = buf.String()

	} else {
		if m, ok := league[positions[0]]; ok {
			result = fmt.Sprintf("#%s \x02%s\x02 - P: %s, GD: %s, Pts: %s", m["position-number"], m["team-name"], m["played"], m["goal-difference"], m["points"])
		}
	}

	return result
}

// Get upcoming games from all leagues
func AllFixtures(zone, input string) string {

	input = strings.TrimSpace(input)
	wantedDate := 1
	if len(input) > 0 {
		var err error
		wantedDate, err = strconv.Atoi(input)
		if err != nil {
			return "Error - " + err.Error()
		}
	}
	if wantedDate < 1 {
		wantedDate = 1
	}

	url := "" //fixtures[zone]
	resp, err := http.Get(url)
	if err != nil {
		return "Error - " + err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Page not found")
		return "Error - Page not found."
	}

	// Run HTML parser to get DOM
	root, err := html.Parse(resp.Body)
	if err != nil {
		return "Error - " + err.Error()
	}

	// Helper function to return all text nodes in a subtree
	var pr func(*html.Node) []string
	pr = func(n *html.Node) []string {
		var result []string

	OUTER:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Skip preview field nodes
			if c.Type == html.ElementNode && c.Data == "td" {
				for _, a := range c.Attr {
					if a.Key == "class" && a.Val == "status" {
						continue OUTER
					}
				}
			}

			// Add text fields
			if c.Type == html.TextNode {
				trimmed := strings.TrimSpace(c.Data)
				if len(trimmed) > 0 {
					result = append(result, trimmed)
				}
			}
			result = append(result, pr(c)...)
		}

		return result
	}

	var output []string

	// Helper function to look through DOM for elements of interest
	var fm func(*html.Node)
	fm = func(n *html.Node) {

		// Handle individual football matches
		if n.Type == html.ElementNode && n.Data == "tr" {
			for _, a := range n.Attr {
				if a.Key == "class" && (a.Val == "report" || a.Val == "preview" || a.Val == "live") {
					match := pr(n)
					if len(match) == 0 {
						break
					}
					match = match[1:]

					// Add bold tags to teams
					if len(match) == 4 {
						match[0] = "\x02" + match[0] + "\x02"
						match[2] = "\x02" + match[2] + "\x02"
					} else if len(match) == 5 {
						match[1] = "\x02" + match[1] + "\x02"
						match[3] = "\x02" + match[3] + "\x02"
					}

					output = append(output, strings.Join(match, " "))
					break
				}
			}
		}
		// Dive into child elements
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			fm(c)
		}
	}

	var gameDate string
	datesSeen := 0

	// Helper function to look through DOM for elements of interest
	var f func(*html.Node)
	f = func(n *html.Node) {

		// Grab very first h2 element
		if n.Type == html.ElementNode && n.Data == "h2" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "table-header" {
					// If we found second instance, we must be done
					datesSeen++
					if datesSeen == wantedDate {
						date := pr(n)
						if len(date) == 0 {
							break
						}
						gameDate = date[0]
					}
					break
				}
			}
		}

		// Dig through table
		if datesSeen == wantedDate && n.Type == html.ElementNode && n.Data == "table" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "table-stats" {
					fm(n)
					break
				}
			}
		}

		// Dive into child elements
		for c := n.FirstChild; c != nil && datesSeen <= wantedDate; c = c.NextSibling {
			f(c)
		}
	}
	f(root)

	if len(output) == 0 {
		fmt.Println("Fixtures not found")
		return "Error - No fixtures found."
	}

	dateRepl := strings.NewReplacer(
		"Monday", "Mon",
		"Tuesday", "Tue",
		"Wednesday", "Wed",
		"Thursday", "Thu",
		"Friday", "Fri",
		"Saturday", "Sat",
		"Sunday", "Sun",
		"January", "Jan",
		"February", "Feb",
		"March", "Mar",
		"April", "Apr",
		"June", "Jun",
		"July", "Jul",
		"August", "Aug",
		"September", "Sep",
		"October", "Oct",
		"November", "Nov",
		"December", "Dec",
		"st ", " ",
		"nd ", " ",
		"rd ", " ",
		"th ", " ",
	)

	matchRepl := strings.NewReplacer(
		"  ", " ",
		"Half time", "HT",
		"Full time", "FT",
	)

	result := dateRepl.Replace(gameDate) + " - " + matchRepl.Replace(strings.Join(output, " | "))
	if len(result) > 500 {
		result = result[:500]
	}

	return result
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

		var result string
		arg := event.Message()[len(cmd):]
		if strings.HasPrefix(arg, "#") {
			result = ShowTable(arg[1:])
		} else {
			result = TablePosition(arg)
		}

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

		result := AllFixtures("AU", event.Message()[len(cmd):])

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
