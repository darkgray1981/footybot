# footybot
Football information dispensing IRC bot

Parses the BBC Push API to get various fixtures, results and table information for football teams.

The BBC Push API takes an internal "morph" URL which refers to a specific piece of data, the URLs we use are:

morph://data/bbc-morph-sport-football-scores-tabbed-teams-model/isApp/false/limit/4/team/%s/version/1.0.6 - Returns the fixtures and results for a team.

morph://data/bbc-morph-sport-football-scores-tabbed-model/isApp/false/limit/12/tournament/%s/version/2.0.0 - Returns the fixtures and results for a competition/tournament.

morph://data/bbc-morph-football-scores-match-list-data/endDate/%s/startDate/%s/tournament/%s/version/2.2.1/withPlayerActions/false - Returns the fixtures for a tournament in a given date range.

morph://data/bbc-morph-sport-football-tables-data/competition/%s/version/1.5.0 - Returns the league table for a given competition/tournament, note for MLS and such like it is all tables in that competition.

morph://data/bbc-morph-sport-football-table-team-model/team/%s/version/1.0.4 - Returns the league table for a given team - the team name matches the "slug" parameter in the JSON response.

Note: To show tables from multi-group leagues (MLS, World Cup, Euros etc) the ShowTable method can take a zone (i.e. mls) and a subZone (i.e. Western Conference) to just show that league