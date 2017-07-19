package main

import (
	"time"
)

// BBC Response structs
type bbcPushResponse struct {
	Moments []struct {
		Topic string `json:"topic"`
		ID int `json:"id"`
		Payload string `json:"payload"`
		PayloadHash string `json:"payloadHash"`
		LastModified time.Time `json:"lastModified"`
	} `json:"moments"`
	Meta struct {
		PollInterval int `json:"poll-interval"`
	} `json:"meta"`
}

type team struct {
	Key string `json:"key"`
	Scores scores `json:"scores"`
	EventOutcome interface{} `json:"eventOutcome"`
	Name teamName `json:"name"`
	Stats interface{} `json:"stats"`
	Players []interface{} `json:"players"`
}

type teamName struct {
	First string `json:"first"`
	Full string `json:"full"`
	Abbreviation string `json:"abbreviation"`
	Last interface{} `json:"last"`
}

type scores struct {
	Score int `json:"score"`
	HalfTime int `json:"halfTime"`
	FullTime int `json:"fullTime"`
	ExtraTime interface{} `json:"extraTime"`
	Shootout interface{} `json:"shootout"`
	Aggregate interface{} `json:"aggregate"`
	AggregateGoalsAway interface{} `json:"aggregateGoalsAway"`
}

type tournamentEvent struct {
	EventKey string `json:"eventKey"`
	StartTime time.Time `json:"startTime"`
	StartTimeInUKHHMM string `json:"startTimeInUKHHMM"`
	MinutesElapsed int `json:"minutesElapsed"`
	MinutesIntoAddedTime int `json:"minutesIntoAddedTime"`
	HomeTeam team `json:"homeTeam"`
	AwayTeam team `json:"awayTeam"`
	Comment interface{} `json:"comment"`
	EventProgress struct {
		Period string `json:"period"`
		Status string `json:"status"`
	} `json:"eventProgress"`
	Href interface{} `json:"href"`
	TemporaryRoundReference struct {
		Name struct {
			First interface{} `json:"first"`
			Full interface{} `json:"full"`
			Abbreviation interface{} `json:"abbreviation"`
		} `json:"name"`
		TempID int `json:"tempId"`
	} `json:"temporaryRoundReference,omitempty"`
}

type teamEvent struct {
	EventKey string `json:"eventKey"`
	StartTime time.Time `json:"startTime"`
	MinutesElapsed int `json:"minutesElapsed"`
	MinutesIntoAddedTime interface{} `json:"minutesIntoAddedTime"`
	EventStatus string `json:"eventStatus"`
	EventStatusNote string `json:"eventStatusNote"`
	EventStatusReason interface{} `json:"eventStatusReason"`
	EventOutcomeType interface{} `json:"eventOutcomeType"`
	EventType string `json:"eventType"`
	SeriesWinner interface{} `json:"seriesWinner"`
	CpsID interface{} `json:"cpsId"`
	CpsLive interface{} `json:"cpsLive"`
	Attendance interface{} `json:"attendance"`
	HomeTeam team `json:"homeTeam"`
	AwayTeam team `json:"awayTeam"`
	EventProgress struct {
		Period string `json:"period"`
		Status string `json:"status"`
	} `json:"eventProgress"`
	Players interface{} `json:"players"`
	Venue struct {
		Name struct {
			Abbreviation string `json:"abbreviation"`
			VideCode string `json:"videCode"`
			First string `json:"first"`
			Full string `json:"full"`
		} `json:"name"`
		HomeCountry interface{} `json:"homeCountry"`
	} `json:"venue"`
	Officials []interface{} `json:"officials"`
	TournamentInfo interface{} `json:"tournamentInfo"`
	StartTimeInUKHHMM string `json:"startTimeInUKHHMM"`
	Comment interface{} `json:"comment"`
	Href interface{} `json:"href"`
	TournamentName struct {
		First string `json:"first"`
		Full string `json:"full"`
		Abbreviation string `json:"abbreviation"`
	} `json:"tournamentName"`
	TournamentSlug string `json:"tournamentSlug"`
	DateString string `json:"dateString"`
}

type tournamentMatch struct {
	TotalEvents int `json:"totalEvents"`
	Date string `json:"date"`
	Tournament struct {
		Name struct {
			First string `json:"first"`
			Full string `json:"full"`
			Abbreviation string `json:"abbreviation"`
		} `json:"name"`
		Slug string `json:"slug"`
		Stages []struct {
			Name interface{} `json:"name"`
			Rounds []struct {
				Name struct {
					First interface{} `json:"first"`
					Full interface{} `json:"full"`
					Abbreviation interface{} `json:"abbreviation"`
				} `json:"name"`
				Events []tournamentEvent `json:"events"`
			} `json:"rounds"`
		} `json:"stages"`
	} `json:"tournament"`
}

type tournamentMatches struct {
	Today tournamentMatch `json:"today"`
	Fixtures tournamentMatch `json:"fixtures"`
	Results tournamentMatch `json:"results"`
}

type teamMatch struct {
	Meta struct {
		ResponseCode int `json:"responseCode"`
		ErrorMessage interface{} `json:"errorMessage"`
		Headers struct {
			ContentType string `json:"content-type"`
		} `json:"headers"`
	} `json:"meta"`
	Body struct {
		Rounds []struct {
			Name struct {
				First        string `json:"first"`
				Full         string `json:"full"`
				Abbreviation string `json:"abbreviation"`
			} `json:"name"`
			Events []teamEvent  `json:"events"`
		} `json:"rounds"`
	} `json:"body"`
}

type teamMatches struct {
	Fixtures teamMatch `json:"fixtures"`
	Today teamMatch `json:"today"`
	Results teamMatch `json:"results"`
}

type fixtureList struct {
	FixtureListMeta struct {
		ScorersButtonShouldBeEnabled bool `json:"scorersButtonShouldBeEnabled"`
	} `json:"fixtureListMeta"`
	MatchData []struct {
		TournamentMeta struct {
			TournamentSlug string `json:"tournamentSlug"`
			TournamentName struct {
				First string `json:"first"`
				Full string `json:"full"`
				Abbreviation string `json:"abbreviation"`
			} `json:"tournamentName"`
		} `json:"tournamentMeta"`
		TournamentDatesWithEvents map[string][]tournamentFixtures `json:"tournamentDatesWithEvents"`
	} `json:"matchData"`
}

type tournamentFixtures struct {
	Round struct {
		Key string `json:"key"`
		Name struct {
			First string `json:"first"`
			Full string `json:"full"`
			Abbreviation string `json:"abbreviation"`
		} `json:"name"`
	} `json:"round"`
	Events []teamEvent  `json:"events"`
}

// Internal structs
type footballMatch struct {
	kickOffTime time.Time
	isFixture bool
	inProgress bool
	isResult bool
	isTournamentGame bool
	Tournament string
	HomeTeam team
	AwayTeam team
	minutesElapsed int
}

type footballMatches struct {
	fixtures []footballMatch
	results []footballMatch
}